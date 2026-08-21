package httpcachex_test

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"example.com/httpcachex"
)

type rtFunc10 func(*http.Request) (*http.Response, error)

func (f rtFunc10) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestBug10_CloseFlushesBeforeDrop(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c.json")
	rt := rtFunc10(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Cache-Control": []string{"max-age=60"}}, Body: io.NopCloser(strings.NewReader("keep")), Request: r}, nil
	})
	c, err := httpcachex.New(httpcachex.Options{Transport: rt, PersistPath: path})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = c.Fetch(httpcachex.Request{Method: "GET", URL: "http://x/keep"})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	c2, err := httpcachex.New(httpcachex.Options{Transport: rt, PersistPath: path})
	if err != nil {
		t.Fatal(err)
	}
	defer c2.Close()
	if err := c2.LoadPersist(); err != nil {
		t.Fatal(err)
	}
	if len(c2.Snapshot()) < 1 {
		t.Fatal("Close flushed empty cache")
	}
}
