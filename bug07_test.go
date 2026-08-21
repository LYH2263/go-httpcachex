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

type rtFunc7 func(*http.Request) (*http.Response, error)

func (f rtFunc7) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestBug07_FetchContextHonorsCancel(t *testing.T) {
	rt := rtFunc7(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Cache-Control": []string{"max-age=60"}}, Body: io.NopCloser(strings.NewReader("c")), Request: r}, nil
	})
	c, err := httpcachex.New(httpcachex.Options{Transport: rt})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err = c.FetchContext(ctx, httpcachex.Request{Method: "GET", URL: "http://x/c"})
	if err == nil || (!errors.Is(err, httpcachex.ErrCanceled) && !errors.Is(err, context.Canceled)) {
		t.Fatalf("got %v", err)
	}
}
