package scenario

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"go.uber.org/zap"

	phttp "github.com/yandex/pandora/components/guns/http"
	"github.com/yandex/pandora/core"
	"github.com/yandex/pandora/core/aggregator/netsample"
)

type Gun interface {
	Shoot(ammo Ammo)
	Bind(sample netsample.Aggregator, deps core.GunDeps) error
}

const (
	EmptyTag = "__EMPTY__"
)

type BaseGun struct {
	DebugLog   bool // Automaticaly set in Bind if Log accepts debug messages.
	Config     phttp.BaseGunConfig
	Connect    func(ctx context.Context) error // Optional hook.
	OnClose    func() error                    // Optional. Called on Close().
	Aggregator netsample.Aggregator            // Lazy set via BindResultTo.
	AnswLog    *zap.Logger
	core.GunDeps
	scheme         string
	hostname       string
	targetResolved string
	client         Client
	templater      *Templater
}

var _ Gun = (*BaseGun)(nil)
var _ io.Closer = (*BaseGun)(nil)

func (b *BaseGun) Bind(aggregator netsample.Aggregator, deps core.GunDeps) error {
	log := deps.Log
	if ent := log.Check(zap.DebugLevel, "Gun bind"); ent != nil {
		// Enable debug level logging during shooting. Creating log entries isn't free.
		b.DebugLog = true
	}

	if b.Aggregator != nil {
		log.Panic("already binded")
	}
	if aggregator == nil {
		log.Panic("nil aggregator")
	}
	b.Aggregator = aggregator
	b.GunDeps = deps

	return nil
}

// Shoot is thread safe iff Do and Connect hooks are thread safe.
func (b *BaseGun) Shoot(ammo Ammo) {
	if b.Aggregator == nil {
		zap.L().Panic("must bind before shoot")
	}
	if b.Connect != nil {
		err := b.Connect(b.Ctx)
		if err != nil {
			b.Log.Warn("Connect fail", zap.Error(err))
			return
		}
	}

	err := b.shoot(ammo)
	if err != nil {
		b.Log.Warn("Invalid ammo", zap.Uint64("request", ammo.ID()), zap.Error(err))
		return
	}
}

func (g *BaseGun) Do(req *http.Request) (*http.Response, error) {

	return g.client.Do(req)
}

func (b *BaseGun) Close() error {
	if b.OnClose != nil {
		return b.OnClose()
	}
	return nil
}

func (b *BaseGun) verboseLogging(resp *http.Response, reqBody, respBody []byte) {
	if resp == nil {
		b.Log.Error("Response is nil")
		return
	}
	fields := make([]zap.Field, 0, 4)
	fields = append(fields, zap.String("URL", resp.Request.URL.String()))
	fields = append(fields, zap.String("Host", resp.Request.Host))
	fields = append(fields, zap.Any("Headers", resp.Request.Header))
	if reqBody != nil {
		fields = append(fields, zap.ByteString("Body", reqBody))
	}
	b.Log.Debug("Request debug info", fields...)

	fields = fields[:0]
	fields = append(fields, zap.Int("Status Code", resp.StatusCode))
	fields = append(fields, zap.String("Status", resp.Status))
	fields = append(fields, zap.Any("Headers", resp.Header))
	if reqBody != nil {
		fields = append(fields, zap.ByteString("Body", respBody))
	}
	b.Log.Debug("Response debug info", fields...)
}

func (b *BaseGun) answLogging(bodyBytes []byte, resp *http.Response, respBytes []byte) {
	msg := fmt.Sprintf("REQUEST:\n%s\n", string(bodyBytes))
	b.AnswLog.Debug(msg)

	var writer bytes.Buffer
	err := resp.Header.Write(&writer)
	if err != nil {
		b.AnswLog.Error("error writing header", zap.Error(err))
		return
	}

	msg = fmt.Sprintf("RESPONSE:\n%s %s\n%s\n%s\n", resp.Proto, resp.Status, writer.String(), string(respBytes))
	b.AnswLog.Debug(msg)
}

func (b *BaseGun) shoot(ammo Ammo) error {
	const op = "base_gun.shoot"

	vs := ammo.VariableStorage()
	//outputParams := ammo.ReturnedParams()
	for _, step := range ammo.Steps() {
		reqParts := RequestParts{
			URL:     step.GetURL(),
			Method:  step.GetMethod(),
			Body:    step.GetBody(),
			Headers: step.GetHeaders(),
		}
		if err := b.templater.Apply(&reqParts, vs); err != nil {
			return fmt.Errorf("%s templater.Apply %w", op, err)
		}
		sample := netsample.Acquire(ammo.Name() + "." + step.GetTag())
		var reader io.Reader
		if reqParts.Body != nil {
			reader = bytes.NewReader(reqParts.Body)
		}

		req, err := http.NewRequest(reqParts.Method, reqParts.URL, reader)
		if err != nil {
			b.reportErr(sample, err)
			return fmt.Errorf("%s http.NewRequest %w", op, err)
		}
		if req.Host == "" {
			req.Host = b.hostname
		}
		req.URL.Host = b.targetResolved
		req.URL.Scheme = b.scheme

		var reqBytes []byte
		if b.Config.AnswLog.Enabled {
			reqBytes, err = httputil.DumpRequestOut(req, true)
			if err != nil {
				b.Log.Error("Error dumping request: %s", zap.Error(err))
			}
		}

		resp, err := b.Do(req)
		if err != nil {
			b.reportErr(sample, err)
			return fmt.Errorf("%s b.Do %w", op, err)
		}
		b.Aggregator.Report(sample)

		var respBody []byte
		if b.Config.AnswLog.Enabled || b.DebugLog || b.templater.needsParseResponse(step.ReturnedParams()) {
			respBody, err = io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("%s io.ReadAll %w", op, err)
			}
		}

		// TODO: is it needed to read body here in every case?
		// For read body we should use io.Copy(io.Discard, resp.Body)

		if b.DebugLog {
			b.verboseLogging(resp, reqBytes, respBody)
		}
		if b.Config.AnswLog.Enabled {
			b.answReqRespLogging(reqBytes, resp, respBody)
		}

		err = b.templater.SaveResponseToVS(resp, "request."+ammo.Name(), step.ReturnedParams(), vs)
		if err != nil {
			return fmt.Errorf("%s templater.SaveResponseToVS %w", op, err)
		}

		if respBody == nil {
			_, err = io.Copy(io.Discard, resp.Body) // Buffers are pooled for ioutil.Discard
			if err != nil {
				return fmt.Errorf("%s io.Copy %w", op, err)
			}
		}
		resp.Body.Close()
	}
	return nil
}

func (b *BaseGun) answReqRespLogging(reqBytes []byte, resp *http.Response, respBytes []byte) {
	switch b.Config.AnswLog.Filter {
	case "all":
		b.answLogging(reqBytes, resp, respBytes)

	case "warning":
		if resp.StatusCode >= 400 {
			b.answLogging(reqBytes, resp, respBytes)
		}

	case "error":
		if resp.StatusCode >= 500 {
			b.answLogging(reqBytes, resp, respBytes)
		}
	}
}

func (b *BaseGun) reportErr(sample *netsample.Sample, err error) {
	if err == nil {
		return
	}
	sample.AddTag(EmptyTag)
	sample.SetProtoCode(0)
	sample.SetErr(err)
	b.Aggregator.Report(sample)
}

func autotag(depth int, URL *url.URL) string {
	path := URL.Path
	var ind int
	for ; ind < len(path); ind++ {
		if path[ind] == '/' {
			if depth == 0 {
				break
			}
			depth--
		}
	}
	return path[:ind]
}
