package policy

import "time"

// Rule5 缓存策略片段。
type Rule5 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule5() Rule5 {
	return Rule5{
		Name:   "rule5",
		MaxAge: time.Duration(5) * time.Second,
		SWR:    time.Duration(5*2) * time.Second,
	}
}

func (r Rule5) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule5) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule5) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule5(a, b Rule5) Rule5 {
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
