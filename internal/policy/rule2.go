package policy

import "time"

// Rule2 缓存策略片段。
type Rule2 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule2() Rule2 {
	return Rule2{
		Name:   "rule2",
		MaxAge: time.Duration(2) * time.Second,
		SWR:    time.Duration(2*2) * time.Second,
	}
}

func (r Rule2) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule2) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule2) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule2(a, b Rule2) Rule2 {
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
