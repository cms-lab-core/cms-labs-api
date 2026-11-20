package http_debug

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestTransportFailureWithoutResponse(t *testing.T) {
	failure := errors.New("connection reset")
	for _, logger := range []*zerolog.Logger{nil, func() *zerolog.Logger { l := zerolog.Nop(); return &l }()} {
		transport := &LoggingTransport{Logger: logger, Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, failure })}
		request, _ := http.NewRequest(http.MethodPost, "https://lms.invalid/token", strings.NewReader("request-body"))
		response, err := transport.RoundTrip(request)
		if response != nil || !errors.Is(err, failure) {
			t.Fatalf("response=%v error=%v", response, err)
		}
	}
}

func TestLoggingDoesNotReadBodiesOrExposeCredentials(t *testing.T) {
	var logs bytes.Buffer
	logger := zerolog.New(&logs).Level(zerolog.DebugLevel)
	request, _ := http.NewRequest(http.MethodPost, "https://lms.invalid/token?private=query-value", strings.NewReader("private-assertion"))
	request.Header.Set("Authorization", "Bearer private-header")
	body := io.NopCloser(strings.NewReader("private-token-response"))
	transport := &LoggingTransport{Logger: &logger, Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		data, err := io.ReadAll(r.Body)
		if err != nil || string(data) != "private-assertion" {
			t.Fatal("request body was changed")
		}
		return &http.Response{StatusCode: 200, Body: body, Header: make(http.Header)}, nil
	})}
	response, err := transport.RoundTrip(request)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(response.Body)
	if err != nil || string(data) != "private-token-response" {
		t.Fatal("response body was consumed")
	}
	response.Body.Close()
	for _, private := range []string{"private-assertion", "private-header", "private-token-response", "query-value"} {
		if strings.Contains(logs.String(), private) {
			t.Fatal("private HTTP data was logged")
		}
	}
	if !strings.Contains(logs.String(), `"status_code":200`) {
		t.Fatal("status metadata missing")
	}
}
