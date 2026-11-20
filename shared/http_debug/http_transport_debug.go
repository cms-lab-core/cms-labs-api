package http_debug

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type LoggingTransport struct {
	*zerolog.Logger
	Transport http.RoundTripper
}

func (s *LoggingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	transport := s.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	started := time.Now()
	response, err := transport.RoundTrip(r)
	if s.Logger != nil {
		status := 0
		if response != nil {
			status = response.StatusCode
		}
		s.Logger.Debug().Str("method", r.Method).Str("host", r.URL.Host).Int("status_code", status).Dur("duration", time.Since(started)).Err(err).Msg("HTTP request completed")
	}
	return response, err
}
