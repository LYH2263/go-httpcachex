package policy

import "time"

// Rule11 缓存策略片段。
type Rule11 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule11() Rule11 {
	return Rule11{
		Name:   "rule11",
		MaxAge: time.Duration(11) * time.Second,
		SWR:    time.Duration(11*2) * time.Second,
	}
}

func (r Rule11) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule11) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule11) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule11(a, b Rule11) Rule11 {
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
