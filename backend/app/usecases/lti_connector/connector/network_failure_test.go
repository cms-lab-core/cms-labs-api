package connector

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestTokenRequestConnectionFailureReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		connection.Close()
	}))
	defer server.Close()
	request, err := http.NewRequest(http.MethodPost, server.URL, strings.NewReader("assertion"))
	if err != nil {
		t.Fatal(err)
	}
	logger := zerolog.Nop()
	_, err = sendRequest(request, &logger)
	if err == nil {
		t.Fatal("closed connection accepted")
	}
	if !errors.Is(err, io.EOF) && !strings.Contains(err.Error(), "connection reset") {
		t.Fatalf("unexpected network error: %v", err)
	}
}
