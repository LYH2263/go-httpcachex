package policy

import "time"

// Rule8 缓存策略片段。
type Rule8 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule8() Rule8 {
	return Rule8{
		Name:   "rule8",
		MaxAge: time.Duration(8) * time.Second,
		SWR:    time.Duration(8*2) * time.Second,
	}
}

func (r Rule8) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule8) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule8) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule8(a, b Rule8) Rule8 {
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
