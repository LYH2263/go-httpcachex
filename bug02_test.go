package httpcachex_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"example.com/httpcachex"
)

type rtFunc2 func(*http.Request) (*http.Response, error)

func (f rtFunc2) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestBug02_VaryKeysSliceAlias(t *testing.T) {
	rt := rtFunc2(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Cache-Control": []string{"max-age=60"}}, Body: io.NopCloser(strings.NewReader("v")), Request: r}, nil
	})
	c, err := httpcachex.New(httpcachex.Options{Transport: rt})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	hdr := map[string][]string{"Vary": []string{"Accept-Language"}, "Accept-Language": []string{"zh"}}
	_, _, err = c.Fetch(httpcachex.Request{Method: "GET", URL: "http://x/v", Headers: hdr})
	if err != nil {
		t.Fatal(err)
	}
	snap := c.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("%d", len(snap))
	}
	snap[0].Vary[0] = "MUT"
	snap2 := c.Snapshot()
	if snap2[0].Vary[0] != "Accept-Language" {
		t.Fatalf("vary aliased: %v", snap2[0].Vary)
	}
}
