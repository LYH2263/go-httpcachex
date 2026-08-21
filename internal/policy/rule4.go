package policy

import "time"

// Rule4 缓存策略片段。
type Rule4 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule4() Rule4 {
	return Rule4{
		Name:   "rule4",
		MaxAge: time.Duration(4) * time.Second,
		SWR:    time.Duration(4*2) * time.Second,
	}
}

func (r Rule4) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule4) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule4) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule4(a, b Rule4) Rule4 {
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
