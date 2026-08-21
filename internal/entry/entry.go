package entry

import "time"

type Entry struct {
	Key        string
	URL        string
	Vary       []string
	Status     int
	Headers    map[string][]string
	Body       []byte
	Hits       int64
	Stored     time.Time
	FreshUntil time.Time
	StaleUntil time.Time
}
