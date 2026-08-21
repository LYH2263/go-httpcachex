package policy

import "time"

// Rule15 缓存策略片段。
type Rule15 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule15() Rule15 {
	return Rule15{
		Name:   "rule15",
		MaxAge: time.Duration(15) * time.Second,
		SWR:    time.Duration(15*2) * time.Second,
	}
}

func (r Rule15) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule15) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule15) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule15(a, b Rule15) Rule15 {
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
