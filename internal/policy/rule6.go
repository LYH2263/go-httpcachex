package policy

import "time"

// Rule6 缓存策略片段。
type Rule6 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule6() Rule6 {
	return Rule6{
		Name:   "rule6",
		MaxAge: time.Duration(6) * time.Second,
		SWR:    time.Duration(6*2) * time.Second,
	}
}

func (r Rule6) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule6) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule6) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule6(a, b Rule6) Rule6 {
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
