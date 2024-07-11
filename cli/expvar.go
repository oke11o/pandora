package cli

import (
	"time"

	"github.com/yandex/pandora/core/engine"
	"github.com/yandex/pandora/lib/monitoring"
	"go.uber.org/zap"
)

func startReport(m engine.Metrics) {
	evReqPS := monitoring.NewCounter("engine_ReqPS")
	evResPS := monitoring.NewCounter("engine_ResPS")
	evActiveUsers := monitoring.NewCounter("engine_ActiveUsers")
	evActiveRequests := monitoring.NewCounter("engine_ActiveRequests")
	evLastMaxActiveRequests := monitoring.NewCounter("engine_LastMaxActiveRequests")
	requests := m.Request.Get()
	responses := m.Response.Get()
	go func() {
		var requestsNew, responsesNew int64
		// TODO(skipor): there is no guarantee, that we will run exactly after 1 second.
		// So, when we get 1 sec +-10ms, we getting 990-1010 calculate intervals and +-2% RPS in reports.
		// Consider using rcrowley/go-metrics.Meter.
		for range time.NewTicker(1 * time.Second).C {
			requestsNew = m.Request.Get()
			responsesNew = m.Response.Get()
			rps := responsesNew - responses
			reqps := requestsNew - requests
			activeUsers := m.InstanceStart.Get() - m.InstanceFinish.Get()
			activeRequests := requestsNew - responsesNew
			lastMaxActiveRequests := int64(m.BusyInstances.Flush())
			zap.S().Infof(
				"[ENGINE] %d resp/s; %d req/s; %d users; %d active\n",
				rps, reqps, activeUsers, lastMaxActiveRequests)

			requests = requestsNew
			responses = responsesNew

			evActiveUsers.Set(activeUsers)
			evActiveRequests.Set(activeRequests)
			evLastMaxActiveRequests.Set(lastMaxActiveRequests)
			evReqPS.Set(reqps)
			evResPS.Set(rps)
		}
	}()
}
