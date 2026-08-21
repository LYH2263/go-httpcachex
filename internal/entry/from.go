package entry

import "time"

func FromHTTP(k, url string, vary []string, status int, headers map[string][]string, body []byte, now time.Time, ttl, swr time.Duration) Entry {
	return Entry{
		Key:        k,
		URL:        url,
		Vary:       append([]string(nil), vary...),
		Status:     status,
		Headers:    CloneHeaders(headers),
		Body:       CloneBytes(body),
		Stored:     now,
		FreshUntil: now.Add(ttl),
		StaleUntil: now.Add(ttl + swr),
	}
}
