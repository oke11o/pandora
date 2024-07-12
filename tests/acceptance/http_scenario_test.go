package acceptance

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/yandex/pandora/core/engine"
	"github.com/yandex/pandora/examples/http/server"
	"github.com/yandex/pandora/lib/testutil"
)

func TestHTTPScenarioSuite(t *testing.T) {
	suite.Run(t, new(HTTPScenarioSuite))
}

type HTTPScenarioSuite struct {
	suite.Suite
	fs      afero.Fs
	log     *zap.Logger
	metrics engine.Metrics
	addr    string
	srv     *server.Server
}

func (s *HTTPScenarioSuite) SetupSuite() {
	s.fs = afero.NewOsFs()
	testOnce.Do(importDependencies(s.fs))

	s.log = testutil.NewNullLogger()
	s.metrics = engine.NewMetrics("http_scenario_suite")

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	port := os.Getenv("PORT") // TODO: how to set free port in CI?
	if port == "" {
		port = "8886"
	}

	s.addr = "localhost:" + port
	s.srv = server.NewServer(s.addr, logger, time.Now().UnixNano())
	s.srv.ServeAsync()

	go func() {
		err := <-s.srv.Err()
		s.NoError(err)
	}()
}

func (s *HTTPScenarioSuite) TearDownSuite() {
	err := s.srv.Shutdown(context.Background())
	s.NoError(err)
}

func (s *HTTPScenarioSuite) SetupTest() {
	s.srv.Stats().Reset()
}

func (s *HTTPScenarioSuite) Test_Http_Check_Passes() {
	tests := []struct {
		name           string
		filecfg        string
		wantErrContain string
		wantCnt        int
		wantStats      *server.Stats
	}{
		{
			name:    "base",
			filecfg: "testdata/http_scenario/scenario.yaml",
			wantCnt: 4,
			wantStats: &server.Stats{
				Auth200:  map[int64]uint64{1: 2, 2: 1, 3: 1},
				List200:  map[int64]uint64{1: 2, 2: 1, 3: 1},
				Order200: map[int64]uint64{1: 6, 2: 3, 3: 3},
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			conf := parseConfigFile(s.T(), tt.filecfg, s.addr)
			s.Require().Equal(1, len(conf.Engine.Pools))
			aggr := &aggregator{}
			conf.Engine.Pools[0].Aggregator = aggr

			pandora := engine.New(s.log, s.metrics, conf.Engine)

			err := pandora.Run(context.Background())
			if tt.wantErrContain != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tt.wantErrContain)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tt.wantStats, s.srv.Stats())
		})
	}
}
