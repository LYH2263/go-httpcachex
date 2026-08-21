package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"example.com/httpcachex"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func main() {
	addr := flag.String("addr", ":8106", "listen")
	web := flag.String("web", "web", "web dir")
	data := flag.String("data", "data", "data dir")
	flag.Parse()
	_ = os.MkdirAll(*data, 0o755)

	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Cache-Control": []string{"max-age=30"}, "Content-Type": []string{"text/plain"}},
			Body:       http.NoBody,
			Request:    r,
		}, nil
	})
	c, err := httpcachex.New(httpcachex.Options{
		Transport: rt, PersistPath: filepath.Join(*data, "cache.json"), AuditPath: filepath.Join(*data, "audit.log"),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	var mu sync.Mutex
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(*web)))
	mux.HandleFunc("/api/entries", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		writeJSON(w, c.Snapshot())
	})
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		writeJSON(w, c.Stats())
	})
	mux.HandleFunc("/api/purge", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST", 405)
			return
		}
		var body struct{ Key string }
		_ = json.NewDecoder(r.Body).Decode(&body)
		mu.Lock()
		err := c.Purge(body.Key)
		mu.Unlock()
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		writeJSON(w, map[string]string{"ok": "1"})
	})
	log.Printf("cached on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
