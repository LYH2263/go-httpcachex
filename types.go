package httpcachex

import "time"

type Request struct {
	Method  string
	URL     string
	Headers map[string][]string
}

type Response struct {
	Status  int
	Headers map[string][]string
	Body    []byte
	Stored  time.Time
	FreshUntil time.Time
	StaleUntil time.Time
}

type EntryView struct {
	Key     string
	URL     string
	Vary    []string
	Status  int
	Hits    int64
	BodyLen int
	FreshUntil time.Time
	StaleUntil time.Time
}
