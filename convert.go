package httpcachex

import (
	"time"

	"example.com/httpcachex/internal/entry"
)

func toResponse(e entry.Entry) Response {
	return Response{
		Status:     e.Status,
		Headers:    entry.CloneHeaders(e.Headers),
		Body:       entry.CloneBytes(e.Body),
		Stored:     e.Stored,
		FreshUntil: e.FreshUntil,
		StaleUntil: e.StaleUntil,
	}
}

func entryFrom(k, url string, vary []string, resp Response, now time.Time, ttl, swr time.Duration) entry.Entry {
	return entry.FromHTTP(k, url, vary, resp.Status, resp.Headers, resp.Body, now, ttl, swr)
}
