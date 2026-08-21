package httpcachex_test

import (
	"errors"
	"testing"

	"example.com/httpcachex"
)

func TestBug04_NilTransportNoPanic(t *testing.T) {
	c, err := httpcachex.New(httpcachex.Options{Transport: nil})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_, _, err = c.Fetch(httpcachex.Request{Method: "GET", URL: "http://x/n"})
	if !errors.Is(err, httpcachex.ErrNoTransport) {
		t.Fatalf("got %v", err)
	}
}
