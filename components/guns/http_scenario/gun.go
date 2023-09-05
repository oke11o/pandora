package httpscenario

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

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
}

var _ Gun = (*BaseGun)(nil)
var _ io.Closer = (*BaseGun)(nil)

func (g *BaseGun) Bind(aggregator netsample.Aggregator, deps core.GunDeps) error {
	log := deps.Log
	if ent := log.Check(zap.DebugLevel, "Gun bind"); ent != nil {
		// Enable debug level logging during shooting. Creating log entries isn't free.
		g.DebugLog = true
	}

	if g.Aggregator != nil {
		log.Panic("already binded")
	}
	if aggregator == nil {
		log.Panic("nil aggregator")
	}
	g.Aggregator = aggregator
	g.GunDeps = deps

	return nil
}

// Shoot is thread safe iff Do and Connect hooks are thread safe.
func (g *BaseGun) Shoot(ammo Ammo) {
	if g.Aggregator == nil {
		zap.L().Panic("must bind before shoot")
	}
	if g.Connect != nil {
		err := g.Connect(g.Ctx)
		if err != nil {
			g.Log.Warn("Connect fail", zap.Error(err))
			return
		}
	}

	templateVars := map[string]any{
		"source": ammo.Sources().Variables(),
	}

	err := g.shoot(ammo, templateVars)
	if err != nil {
		g.Log.Warn("Invalid ammo", zap.Uint64("request", ammo.ID()), zap.Error(err))
		return
	}
}

func (g *BaseGun) Do(req *http.Request) (*http.Response, error) {
	return g.client.Do(req)
}

func (g *BaseGun) Close() error {
	if g.OnClose != nil {
		return g.OnClose()
	}
	return nil
}

func (g *BaseGun) verboseLogging(resp *http.Response, reqBody, respBody []byte) {
	if resp == nil {
		g.Log.Error("Response is nil")
		return
	}
	fields := make([]zap.Field, 0, 4)
	fields = append(fields, zap.String("URL", resp.Request.URL.String()))
	fields = append(fields, zap.String("Host", resp.Request.Host))
	fields = append(fields, zap.Any("Headers", resp.Request.Header))
	if reqBody != nil {
		fields = append(fields, zap.ByteString("Body", reqBody))
	}
	g.Log.Debug("Request debug info", fields...)

	fields = fields[:0]
	fields = append(fields, zap.Int("Status Code", resp.StatusCode))
	fields = append(fields, zap.String("Status", resp.Status))
	fields = append(fields, zap.Any("Headers", resp.Header))
	if reqBody != nil {
		fields = append(fields, zap.ByteString("Body", respBody))
	}
	g.Log.Debug("Response debug info", fields...)
}

func (g *BaseGun) answLogging(bodyBytes []byte, resp *http.Response, respBytes []byte, stepName string) {
	msg := fmt.Sprintf("REQUEST[%s]:\n%s\n", stepName, string(bodyBytes))
	g.AnswLog.Debug(msg)

	var writer bytes.Buffer
	err := resp.Header.Write(&writer)
	if err != nil {
		g.AnswLog.Error("error writing header", zap.Error(err))
		return
	}

	msg = fmt.Sprintf("RESPONSE[%s]:\n%s %s\n%s\n%s\n", stepName, resp.Proto, resp.Status, writer.String(), string(respBytes))
	g.AnswLog.Debug(msg)
}

func (g *BaseGun) shoot(ammo Ammo, templateVars map[string]any) error {
	const op = "base_gun.shoot"
	if templateVars == nil {
		templateVars = map[string]any{}
	}

	requestVars := map[string]any{}
	templateVars["request"] = requestVars

	startAt := time.Now()
	stepID := strings.Builder{}
	rnd := rand.Int()
	for _, step := range ammo.Steps() {
		// 1. Построить stepID
		if g.Config.AnswLog.Enabled {
			stepID.WriteString(ammo.Name())
			stepID.WriteByte('.')
			stepID.WriteString(strconv.Itoa(rnd))
			stepID.WriteByte('.')
			stepID.WriteString(strconv.Itoa(int(ammo.ID())))
			stepID.WriteByte('.')
			stepID.WriteString(step.GetName())
		}

		// 2. Подготовить stepVars
		stepVars := map[string]any{}
		requestVars[step.GetName()] = stepVars

		// 3. Выполнить preprocessor
		preProcessor := step.Preprocessor()
		if preProcessor != nil {
			preProcVars, err := preProcessor.Process(templateVars)
			if err != nil {
				return fmt.Errorf("%s preProcessor %w", op, err)
			}
			stepVars["preprocessor"] = preProcVars
			if g.DebugLog {
				g.GunDeps.Log.Debug("Preprocessor variables", zap.Any(fmt.Sprintf(".resuest.%s.preprocessor", step.GetName()), preProcVars))
			}
		}

		reqParts := RequestParts{
			URL:     step.GetURL(),
			Method:  step.GetMethod(),
			Body:    step.GetBody(),
			Headers: step.GetHeaders(),
		}
		sample := netsample.Acquire(ammo.Name() + "." + step.GetTag())

		// 4. Выполнить templater
		templater := step.GetTemplater()
		if err := templater.Apply(&reqParts, templateVars, ammo.Name(), step.GetName()); err != nil {
			g.reportErr(sample, err)
			return fmt.Errorf("%s templater.Apply %w", op, err)
		}

		// 5. Подготовить request
		var reader io.Reader
		if reqParts.Body != nil {
			reader = bytes.NewReader(reqParts.Body)
		}

		req, err := http.NewRequest(reqParts.Method, reqParts.URL, reader)
		if err != nil {
			g.reportErr(sample, err)
			return fmt.Errorf("%s http.NewRequest %w", op, err)
		}
		for k, v := range reqParts.Headers {
			req.Header.Set(k, v)
		}
		if req.Host == "" {
			req.Host = g.hostname
		}
		req.URL.Host = g.targetResolved
		req.URL.Scheme = g.scheme

		var reqBytes []byte
		if g.Config.AnswLog.Enabled {
			reqBytes, err = httputil.DumpRequestOut(req, true)
			if err != nil {
				g.Log.Error("Error dumping request: %s", zap.Error(err))
			}
		}

		timings, req := g.initTracing(req, sample)

		resp, err := g.Do(req)

		g.saveTrace(timings, sample, resp)

		if err != nil {
			g.reportErr(sample, err)
			return fmt.Errorf("%s g.Do %w", op, err)
		}

		processors := step.GetPostProcessors()
		var respBody *bytes.Reader
		var respBodyBytes []byte
		if g.Config.AnswLog.Enabled || g.DebugLog || len(processors) > 0 {
			respBodyBytes, err = io.ReadAll(resp.Body)
			if err == nil {
				respBody = bytes.NewReader(respBodyBytes)
			}
		} else {
			_, err = io.Copy(io.Discard, resp.Body)
		}
		if err != nil {
			return fmt.Errorf("%s io.Copy %w", op, err)
		}
		defer func() {
			err = resp.Body.Close()
			if err != nil {
				g.GunDeps.Log.Error("resp.Body.Close", zap.Error(err))
			}
		}()

		if g.DebugLog {
			g.verboseLogging(resp, reqBytes, respBodyBytes)
		}
		if g.Config.AnswLog.Enabled {
			g.answReqRespLogging(reqBytes, resp, respBodyBytes, stepID.String())
			stepID.Reset()
		}

		postprocessorVars := map[string]any{}
		for _, postprocessor := range processors {
			vars, err := postprocessor.Process(resp, respBody)
			if err != nil {
				g.reportErr(sample, err)
				return fmt.Errorf("%s postprocessor.Postprocess %w", op, err)
			}
			for k, v := range vars {
				postprocessorVars[k] = v
			}
			_, err = respBody.Seek(0, io.SeekStart)
			if err != nil {
				g.reportErr(sample, err)
				return fmt.Errorf("%s postprocessor.Postprocess %w", op, err)
			}
		}
		stepVars["postprocessor"] = postprocessorVars

		sample.SetProtoCode(resp.StatusCode)
		g.Aggregator.Report(sample)

		if g.DebugLog {
			g.GunDeps.Log.Debug("Postprocessor variables", zap.Any(fmt.Sprintf(".resuest.%s.postprocessor", step.GetName()), postprocessorVars))
		}

		if step.GetSleep() > 0 {
			time.Sleep(step.GetSleep())
		}
	}
	spent := time.Since(startAt)
	if ammo.GetMinWaitingTime() > spent {
		time.Sleep(ammo.GetMinWaitingTime() - spent)
	}
	return nil
}

func (g *BaseGun) initTracing(req *http.Request, sample *netsample.Sample) (*phttp.TraceTimings, *http.Request) {
	var timings *phttp.TraceTimings
	if g.Config.HTTPTrace.TraceEnabled {
		var clientTracer *httptrace.ClientTrace
		clientTracer, timings = phttp.CreateHTTPTrace()
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), clientTracer))
	}
	if g.Config.HTTPTrace.DumpEnabled {
		requestDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			g.Log.Error("DumpRequest error", zap.Error(err))
		} else {
			sample.SetRequestBytes(len(requestDump))
		}
	}
	return timings, req
}

func (g *BaseGun) saveTrace(timings *phttp.TraceTimings, sample *netsample.Sample, resp *http.Response) {
	if g.Config.HTTPTrace.TraceEnabled && timings != nil {
		sample.SetReceiveTime(timings.GetReceiveTime())
	}
	if g.Config.HTTPTrace.DumpEnabled && resp != nil {
		responseDump, e := httputil.DumpResponse(resp, true)
		if e != nil {
			g.Log.Error("DumpResponse error", zap.Error(e))
		} else {
			sample.SetResponseBytes(len(responseDump))
		}
	}
	if g.Config.HTTPTrace.TraceEnabled && timings != nil {
		sample.SetConnectTime(timings.GetConnectTime())
		sample.SetSendTime(timings.GetSendTime())
		sample.SetLatency(timings.GetLatency())
	}
}

func (g *BaseGun) answReqRespLogging(reqBytes []byte, resp *http.Response, respBytes []byte, stepName string) {
	switch g.Config.AnswLog.Filter {
	case "all":
		g.answLogging(reqBytes, resp, respBytes, stepName)
	case "warning":
		if resp.StatusCode >= 400 {
			g.answLogging(reqBytes, resp, respBytes, stepName)
		}
	case "error":
		if resp.StatusCode >= 500 {
			g.answLogging(reqBytes, resp, respBytes, stepName)
		}
	}
}

func (g *BaseGun) reportErr(sample *netsample.Sample, err error) {
	if err == nil {
		return
	}
	sample.AddTag(EmptyTag)
	sample.SetProtoCode(0)
	sample.SetErr(err)
	g.Aggregator.Report(sample)
}

func (g *BaseGun) resolveTemplater(templaterType string) (Templater, error) {
	//switch templaterType {
	//case "text":
	//	return NewTextTempalter(), nil
	//case "html":
	//	return NewHTMLTemplater(), nil
	//}
	return nil, nil
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
