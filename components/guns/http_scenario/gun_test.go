package httpscenario

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	phttp "github.com/yandex/pandora/components/guns/http"
	"github.com/yandex/pandora/core"
	"github.com/yandex/pandora/core/aggregator/netsample"
)

type mockSource struct {
	variables map[string]any
}

func (m mockSource) Variables() map[string]any {
	return m.variables
}

func TestBaseGun_shoot(t *testing.T) {
	type fields struct {
		DebugLog       bool
		Config         phttp.BaseGunConfig
		Connect        func(ctx context.Context) error
		OnClose        func() error
		Aggregator     netsample.Aggregator
		AnswLog        *zap.Logger
		GunDeps        core.GunDeps
		scheme         string
		hostname       string
		targetResolved string
		client         Client
		templater      *TextTemplater
	}

	tests := []struct {
		name       string
		stepMocks  []func(t *testing.T, m *MockStep)
		ammoMock   func(t *testing.T, m *MockAmmo)
		clientMock func(t *testing.T, m *MockClient)
		mockSource mockSource
		fields     fields
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "default",
			stepMocks: []func(t *testing.T, m *MockStep){
				func(t *testing.T, step *MockStep) {
					//prepoc := NewMockPreprocessor(t)
					step.On("Preprocessor").Return(nil).Times(1)
					step.On("GetURL").Return("http://localhost:8080").Times(1)
					step.On("GetMethod").Return("GET").Times(1)
					step.On("GetBody").Return(nil).Times(1)
					step.On("GetHeaders").Return(map[string]string{"Content-Type": "application/json"}).Times(1)

					step.On("GetTag").Return("tag").Times(1)
					step.On("GetTemplater").Return("text").Times(1)
					step.On("GetName").Return("step 1").Times(2)

					//postprocessor1 := NewMockPostprocessor(t)
					//postprocessor2 := NewMockPostprocessor(t)
					//postprocessors := []Postprocessor{postprocessor1, postprocessor2}
					step.On("GetPostProcessors").Return(nil).Times(1)
					step.On("GetSleep").Return(time.Duration(0)).Times(1)
				},
				func(t *testing.T, step *MockStep) {
					step.On("Preprocessor").Return(nil).Times(1)
					step.On("GetURL").Return("http://localhost:8080").Times(1)
					step.On("GetMethod").Return("POST").Times(1)
					step.On("GetBody").Return(nil).Times(1)
					step.On("GetHeaders").Return(map[string]string{"Content-Type": "application/json"}).Times(1)

					step.On("GetTag").Return("tag").Times(1)
					step.On("GetTemplater").Return("text").Times(1)
					step.On("GetName").Return("step 1").Times(2)

					step.On("GetPostProcessors").Return(nil).Times(1)
					step.On("GetSleep").Return(time.Duration(0)).Times(1)
				},
			},
			ammoMock: func(t *testing.T, ammo *MockAmmo) {
				ammo.On("Name").Return("testAmmo").Times(4)
				ammo.On("GetMinWaitingTime").Return(time.Duration(0))
			},
			clientMock: func(t *testing.T, client *MockClient) {
				body := io.NopCloser(strings.NewReader("test response body"))
				resp := &http.Response{Body: body}
				client.On("Do", mock.Anything).Return(resp, nil) //TODO: check response after template
			},
			mockSource: mockSource{variables: map[string]any{}},
			wantErr:    assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps := make([]Step, 0, len(tt.stepMocks))
			for _, step := range tt.stepMocks {
				st := NewMockStep(t)
				step(t, st)
				steps = append(steps, st)
			}

			ammo := NewMockAmmo(t)
			ammo.On("Steps").Return(steps)
			ammo.On("Sources").Return(tt.mockSource)
			tt.ammoMock(t, ammo)

			client := NewMockClient(t)
			tt.clientMock(t, client)

			aggregator := netsample.NewMockAggregator(t)
			aggregator.On("Report", mock.Anything)

			g := &BaseGun{Aggregator: aggregator, client: client}
			tt.wantErr(t, g.shoot(ammo), fmt.Sprintf("shoot(%v)", ammo))
		})
	}
}
