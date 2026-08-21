package policy

import "time"

// Rule16 缓存策略片段。
type Rule16 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule16() Rule16 {
	return Rule16{
		Name:   "rule16",
		MaxAge: time.Duration(16) * time.Second,
		SWR:    time.Duration(16*2) * time.Second,
	}
}

func (r Rule16) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule16) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule16) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule16(a, b Rule16) Rule16 {
	out := a
	if b.MaxAge > 0 && (a.MaxAge == 0 || b.MaxAge < a.MaxAge) {
		out.MaxAge = b.MaxAge
	}
	if b.SWR > out.SWR {
		out.SWR = b.SWR
	}
	out.NoStore = a.NoStore || b.NoStore
	out.MustRevalidate = a.MustRevalidate || b.MustRevalidate
	return out
}
