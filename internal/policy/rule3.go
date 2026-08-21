package policy

import "time"

// Rule3 缓存策略片段。
type Rule3 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule3() Rule3 {
	return Rule3{
		Name:   "rule3",
		MaxAge: time.Duration(3) * time.Second,
		SWR:    time.Duration(3*2) * time.Second,
	}
}

func (r Rule3) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule3) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule3) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule3(a, b Rule3) Rule3 {
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
