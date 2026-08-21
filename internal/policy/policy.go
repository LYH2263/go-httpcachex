package policy

import (
	"strings"
	"time"
)

type Policy struct {
	DefaultTTL time.Duration
	SWR        time.Duration
}

func Default() Policy {
	return Policy{DefaultTTL: 30 * time.Second, SWR: 10 * time.Second}
}

func (p Policy) CanStore(h map[string][]string) bool {
	cc := strings.Join(h["Cache-Control"], ",")
	if strings.Contains(cc, "no-store") {
		return false
	}
	if strings.Contains(cc, "private") {
		return false
	}
	return true
}
