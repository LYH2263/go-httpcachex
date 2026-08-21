package httpcachex

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// stubRT serves a fixed cacheable 200 text/html body.
type stubRT struct {
	body []byte
}

func (s *stubRT) RoundTrip(*http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", "text/html")
	_, _ = rec.Write(s.body)
	return rec.Result(), nil
}

func newAliasTestCache(t *testing.T) *Cache {
	t.Helper()
	c, err := New(Options{
		Transport:   &stubRT{body: []byte("hello")},
		DefaultTTL:  time.Hour,
		SWR:         time.Hour,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

// TestHitReturnBodyNotAliased reproduces the aliasing bug: after a cache hit,
// mutating the returned Response.Body must not corrupt the stored entry, so a
// subsequent Fetch of the same URL returns the original body. Before the fix,
// CloneBytes returned src, so Response.Body shared the cached entry's backing
// array and a caller write corrupted the page other tenants read.
func TestHitReturnBodyNotAliased(t *testing.T) {
	c := newAliasTestCache(t)
	req := Request{Method: http.MethodGet, URL: "http://example.test/x"}

	// First fetch populates the cache from upstream (miss).
	first, hit, err := c.Fetch(req)
	if err != nil {
		t.Fatalf("first fetch: %v", err)
	}
	if hit {
		t.Fatalf("first fetch should miss, got hit")
	}
	if !bytes.Equal(first.Body, []byte("hello")) {
		t.Fatalf("first body = %q, want %q", first.Body, "hello")
	}

	// Second fetch must hit and return the same body.
	second, hit, err := c.Fetch(req)
	if err != nil {
		t.Fatalf("second fetch: %v", err)
	}
	if !hit {
		t.Fatalf("second fetch should hit, got miss")
	}
	if !bytes.Equal(second.Body, []byte("hello")) {
		t.Fatalf("second body = %q, want %q", second.Body, "hello")
	}

	// Mutate the returned body. Before the fix this wrote through to the
	// stored entry's backing array, corrupting the cache.
	second.Body[0] = 'X'

	// A third fetch must still return the original, uncorrupted body.
	third, hit, err := c.Fetch(req)
	if err != nil {
		t.Fatalf("third fetch: %v", err)
	}
	if !hit {
		t.Fatalf("third fetch should hit, got miss")
	}
	if !bytes.Equal(third.Body, []byte("hello")) {
		t.Fatalf("cached body mutated by caller: got %q, want %q (aliasing bug)", third.Body, "hello")
	}
}
