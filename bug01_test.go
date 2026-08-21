package httpcachex_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"example.com/httpcachex"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestBug01_CachedBodySliceAlias(t *testing.T) {
	rt := rtFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Cache-Control": []string{"max-age=60"}}, Body: io.NopCloser(strings.NewReader("HELLO")), Request: r}, nil
	})
	c, err := httpcachex.New(httpcachex.Options{Transport: rt})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	resp, _, err := c.Fetch(httpcachex.Request{Method: "GET", URL: "http://x/a"})
	if err != nil {
		t.Fatal(err)
	}
	resp.Body[0] = 'X'
	resp2, hit, err := c.Fetch(httpcachex.Request{Method: "GET", URL: "http://x/a"})
	if err != nil || !hit {
		t.Fatalf("err=%v hit=%v", err, hit)
	}
	if string(resp2.Body) != "HELLO" {
		t.Fatalf("body aliased: %q", resp2.Body)
	}
}
