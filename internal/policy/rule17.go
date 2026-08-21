package policy

import "time"

// Rule17 缓存策略片段。
type Rule17 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule17() Rule17 {
	return Rule17{
		Name:   "rule17",
		MaxAge: time.Duration(17) * time.Second,
		SWR:    time.Duration(17*2) * time.Second,
	}
}

func (r Rule17) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule17) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule17) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule17(a, b Rule17) Rule17 {
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
