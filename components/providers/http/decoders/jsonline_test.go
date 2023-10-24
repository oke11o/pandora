package decoders

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex/pandora/components/providers/http/config"
	"github.com/yandex/pandora/components/providers/http/decoders/ammo"
)

const (
	jsonlineDecoderInput = `{"host": "ya.net", "method": "GET", "uri": "/?sleep=100", "tag": "sleep1", "headers": {"User-agent": "Tank", "Connection": "close"}}
{"host": "ya.net", "method": "POST", "uri": "/?sleep=200", "tag": "sleep2", "headers": {"User-agent": "Tank", "Connection": "close"}, "body": "body_data"}
{"host": "ya.net", "method": "PUT", "uri": "/", "tag": "sleep3", "headers": {"User-agent": "Tank", "Connection": "close"}, "body": "body_data"}


`

	jsonlineDecoderMultiInput = `

{
    "host": "ya.net",
    "method": "GET",
    "uri": "/?sleep=100",
    "tag": "sleep1",
    "headers": {
        "User-agent": "Tank",
        "Connection": "close"
    }
}
{
    "host": "ya.net",
    "method": "POST",
    "uri": "/?sleep=200",
    "tag": "sleep2",
    "headers": {
        "User-agent": "Tank",
        "Connection": "close"
    },
    "body": "body_data"
}

{
    "host": "ya.net",
    "method": "PUT",
    "uri": "/",
    "tag": "sleep3",
    "headers": {
        "User-agent": "Tank",
        "Connection": "close"
    },
    "body": "body_data"
}

`

	jsonlineDecoderArrayInput = `

[
    {
        "host": "ya.net",
        "method": "GET",
        "uri": "/?sleep=100",
        "tag": "sleep1",
        "headers": {
            "User-agent": "Tank",
            "Connection": "close"
        }
    },
    {
        "host": "ya.net",
        "method": "POST",
        "uri": "/?sleep=200",
        "tag": "sleep2",
        "headers": {
            "User-agent": "Tank",
            "Connection": "close"
        },
        "body": "body_data"
    },
    {
        "host": "ya.net",
        "method": "PUT",
        "uri": "/",
        "tag": "sleep3",
        "headers": {
            "User-agent": "Tank",
            "Connection": "close"
        },
        "body": "body_data"
    }
]


`
)

func getJsonlineAmmoWants(t *testing.T) []DecodedAmmo {
	var mustNewAmmo = func(t *testing.T, method string, url string, body []byte, header http.Header, tag string) *ammo.Ammo {
		a := ammo.Ammo{}
		err := a.Setup(method, url, body, header, tag)
		require.NoError(t, err)
		return &a
	}
	return []DecodedAmmo{
		mustNewAmmo(t,
			"GET",
			"http://ya.net/?sleep=100",
			nil,
			http.Header{"Connection": []string{"close"}, "Content-Type": []string{"application/json"}, "User-Agent": []string{"Tank"}},
			"sleep1",
		),
		mustNewAmmo(t,
			"POST",
			"http://ya.net/?sleep=200",
			[]byte("body_data"),
			http.Header{"Connection": []string{"close"}, "Content-Type": []string{"application/json"}, "User-Agent": []string{"Tank"}},
			"sleep2",
		),
		mustNewAmmo(t,
			"PUT",
			"http://ya.net/",
			[]byte("body_data"),
			http.Header{"Connection": []string{"close"}, "Content-Type": []string{"application/json"}, "User-Agent": []string{"Tank"}},
			"sleep3",
		),
	}
}

func Test_jsonlineDecoder_Scan(t *testing.T) {
	cases := []struct {
		name  string
		input string
		wants []DecodedAmmo
	}{
		{
			name:  "default",
			input: jsonlineDecoderInput,
			wants: getJsonlineAmmoWants(t),
		},
		{
			name:  "multiline json",
			input: jsonlineDecoderMultiInput,
			wants: getJsonlineAmmoWants(t),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			decoder, _, err := newJsonlineDecoder(strings.NewReader(tt.input), config.Config{
				Limit: 6,
			}, http.Header{"Content-Type": []string{"application/json"}})
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			for j := 0; j < 2; j++ {
				for i, want := range tt.wants {
					ammo, err := decoder.Scan(ctx)
					assert.NoError(t, err, "iteration %d-%d", j, i)
					assert.Equal(t, want, ammo, "iteration %d-%d", j, i)
				}
			}

			_, err = decoder.Scan(ctx)
			assert.Equal(t, err, ErrAmmoLimit)
			assert.Equal(t, decoder.ammoNum, uint(len(tt.wants)*2))
			assert.Equal(t, decoder.passNum, uint(1))
		})
	}
}
func Test_jsonlineDecoder_readArray(t *testing.T) {
	decoder, _, err := newJsonlineDecoder(strings.NewReader(jsonlineDecoderArrayInput), config.Config{
		Limit: 6,
	}, http.Header{"Content-Type": []string{"application/json"}})
	require.NoError(t, err)

	ammos, err := decoder.readArray(context.Background())
	require.NoError(t, err)
	want := getJsonlineAmmoWants(t)
	require.Equal(t, want, ammos)
}

func Test_jsonlineDecoder_LoadAmmo(t *testing.T) {
	decoder, _, err := newJsonlineDecoder(strings.NewReader(jsonlineDecoderInput), config.Config{
		Limit: 7,
	}, http.Header{"Content-Type": []string{"application/json"}})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	wants := getJsonlineAmmoWants(t)

	ammos, err := decoder.LoadAmmo(ctx)
	assert.NoError(t, err)
	assert.Equal(t, wants, ammos)
	assert.Equal(t, decoder.config.Limit, uint(7))
	assert.Equal(t, decoder.config.Passes, uint(0))
}

func BenchmarkScan_line(b *testing.B) {
	decoder, _, err := newJsonlineDecoder(
		strings.NewReader(jsonlineDecoderInput), config.Config{},
		http.Header{"Content-Type": []string{"application/json"}},
	)
	require.NoError(b, err)

	// Подготовьте контекст для тестирования.
	ctx := context.Background()

	// Запустите бенчмарк.
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decoder.Scan(ctx)
		require.NoError(b, err)
	}
}

func BenchmarkScan_multi(b *testing.B) {
	decoder, _, err := newJsonlineDecoder(
		strings.NewReader(jsonlineDecoderInput), config.Config{},
		http.Header{"Content-Type": []string{"application/json"}},
	)
	require.NoError(b, err)

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decoder.Scan(ctx)
		require.NoError(b, err)
	}
}
