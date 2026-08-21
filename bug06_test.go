package httpcachex_test

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"example.com/httpcachex"
)

type rtFunc6 func(*http.Request) (*http.Response, error)

func (f rtFunc6) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestBug06_PersistFailureNotStored(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "file")
	if err := os.WriteFile(bad, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	rt := rtFunc6(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Cache-Control": []string{"max-age=60"}}, Body: io.NopCloser(strings.NewReader("p")), Request: r}, nil
	})
	c, err := httpcachex.New(httpcachex.Options{Transport: rt, PersistPath: filepath.Join(bad, "c.json")})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_, _, err = c.Fetch(httpcachex.Request{Method: "GET", URL: "http://x/p"})
	if err == nil {
		t.Fatal("expected persist error")
	}
	if len(c.Snapshot()) != 0 {
		t.Fatalf("entry remained: %d", len(c.Snapshot()))
	}
}
