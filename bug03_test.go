package httpcachex_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"example.com/httpcachex"
)

type rtFunc3 func(*http.Request) (*http.Response, error)

func (f rtFunc3) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestBug03_FetchAfterCloseNoPanic(t *testing.T) {
	rt := rtFunc3(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Cache-Control": []string{"max-age=60"}}, Body: io.NopCloser(strings.NewReader("z")), Request: r}, nil
	})
	c, err := httpcachex.New(httpcachex.Options{Transport: rt})
	if err != nil {
		t.Fatal(err)
	}
	_ = c.Close()
	_, _, err = c.Fetch(httpcachex.Request{Method: "GET", URL: "http://x/z"})
	if !errors.Is(err, httpcachex.ErrClosed) {
		t.Fatalf("got %v", err)
	}
}
