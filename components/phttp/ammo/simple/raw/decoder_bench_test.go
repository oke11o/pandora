package raw

import (
	"testing"

	"github.com/yandex/pandora/components/phttp/ammo/simple"
)

var (
	benchTestConfigHeaders = []string{
		"[Host: yourhost.tld]",
		"[Sometest: someval]",
	}
)

const (
	benchTestRequest = "GET / HTTP/1.1\r\n" +
		"Host: yourhost.tld" +
		"Content-Length: 0\r\n" +
		"\r\n"
)

// BenchmarkRawDecoder-4              	  500000	      2040 ns/op	    5152 B/op	      11 allocs/op
// BenchmarkRawDecoderWithHeaders-4   	 1000000	      1944 ns/op	    5168 B/op	      12 allocs/op

func BenchmarkRawDecoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = decodeRequest([]byte(benchTestRequest))
	}
}

func BenchmarkRawDecoderWithHeaders(b *testing.B) {
	b.StopTimer()
	decodedHTTPConfigHeaders, _ := simple.DecodeHTTPConfigHeaders(benchTestConfigHeaders)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		req, _ := decodeRequest([]byte(benchTestRequest))
		simple.EnrichRequestWithHeaders(req, decodedHTTPConfigHeaders)
	}
}
