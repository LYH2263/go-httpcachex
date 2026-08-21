package httpcachex_test

import (
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"example.com/httpcachex"
)

type closeCount struct {
	io.ReadCloser
	n *int32
}

func (c closeCount) Close() error {
	atomic.AddInt32(c.n, 1)
	return c.ReadCloser.Close()
}

type rtFunc9 func(*http.Request) (*http.Response, error)

func (f rtFunc9) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestBug09_UpstreamBodyClosed(t *testing.T) {
	var n int32
	rt := rtFunc9(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Cache-Control": []string{"max-age=60"}},
			Body:       closeCount{ReadCloser: io.NopCloser(strings.NewReader("body")), n: &n},
			Request:    r,
		}, nil
	})
	c, err := httpcachex.New(httpcachex.Options{Transport: rt})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_, _, err = c.Fetch(httpcachex.Request{Method: "GET", URL: "http://x/b"})
	if err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&n) < 1 {
		t.Fatal("upstream body not closed")
	}
}
