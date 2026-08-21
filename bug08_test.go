package httpcachex_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"example.com/httpcachex"
)

type rtFunc8 func(*http.Request) (*http.Response, error)

func (f rtFunc8) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestBug08_RevalidateHonorsContext(t *testing.T) {
	rt := rtFunc8(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Cache-Control": []string{"max-age=60"}}, Body: io.NopCloser(strings.NewReader("r")), Request: r}, nil
	})
	c, err := httpcachex.New(httpcachex.Options{Transport: rt})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_, _, err = c.Fetch(httpcachex.Request{Method: "GET", URL: "http://x/r"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = c.Revalidate(ctx, "GET|http://x/r", httpcachex.Request{Method: "GET", URL: "http://x/r"})
	if err == nil || (!errors.Is(err, httpcachex.ErrCanceled) && !errors.Is(err, context.Canceled)) {
		t.Fatalf("got %v", err)
	}
}
