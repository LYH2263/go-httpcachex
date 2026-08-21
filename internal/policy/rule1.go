package policy

import "time"

// Rule1 缓存策略片段。
type Rule1 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule1() Rule1 {
	return Rule1{
		Name:   "rule1",
		MaxAge: time.Duration(1) * time.Second,
		SWR:    time.Duration(1*2) * time.Second,
	}
}

func (r Rule1) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule1) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule1) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule1(a, b Rule1) Rule1 {
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
