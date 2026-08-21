package httpcachex_test

import (
	"errors"
	"net/http"
	"testing"

	"example.com/httpcachex"
)

type rtErr func(*http.Request) (*http.Response, error)

func (f rtErr) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestBug05_FetchErrorWrapsSentinel(t *testing.T) {
	rt := rtErr(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("boom")
	})
	c, err := httpcachex.New(httpcachex.Options{Transport: rt})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_, _, err = c.Fetch(httpcachex.Request{Method: "GET", URL: "http://x/e"})
	if !errors.Is(err, httpcachex.ErrUpstream) {
		t.Fatalf("got %v", err)
	}
}
