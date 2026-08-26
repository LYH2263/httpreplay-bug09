package httpreplay

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type staticTripper struct {
	status int
	body   string
}

func (s staticTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: s.status,
		Body:       io.NopCloser(strings.NewReader(s.body)),
		Header:     make(http.Header),
	}, nil
}

func TestBug09_ReplayOnceCtxPropagation(t *testing.T) {
	c, _ := OpenCassette(Options{Name: "demo"})
	rec := NewRecorder(c, Options{RealTransport: staticTripper{status: 200, body: "ok"}})
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	if _, err := rec.RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	p := NewPlayer(c, Options{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	live, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	_, err := p.ReplayOnce(ctx, live)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled from ReplayOnce ctx, got %v", err)
	}
}
