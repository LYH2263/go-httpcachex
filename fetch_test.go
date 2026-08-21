package httpcachex

import (
	"errors"
	"net/http"
	"testing"

	"example.com/httpcachex/internal/policy"
)

// New with Transport set to nil must not panic on Fetch. Previously the nil
// RoundTripper reached upstream.Do, where rt.RoundTrip(nil) panicked, taking
// down health checks and blocking rolls.
func TestFetchNilTransportReturnsErrNoTransport(t *testing.T) {
	c, err := New(Options{
		Transport: nil,
		Policy:    policy.Default(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, _, err = c.Fetch(Request{Method: "GET", URL: "https://example.com/"})
	if !errors.Is(err, ErrNoTransport) {
		t.Fatalf("want ErrNoTransport, got %v", err)
	}
}

// A configured transport is unaffected by the nil guard: the request must
// still be attempted against upstream (here it fails in the transport, not
// at the guard).
type boomRT struct{}

func (boomRT) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, errors.New("boom")
}

func TestFetchNilGuardDoesNotShadowRealTransport(t *testing.T) {
	c, err := New(Options{
		Transport: boomRT{},
		Policy:    policy.Default(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, _, err = c.Fetch(Request{Method: "GET", URL: "https://example.com/"})
	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("want ErrUpstream wrapping the transport error, got %v", err)
	}
}
