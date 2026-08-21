package policy

import "time"

// Rule12 缓存策略片段。
type Rule12 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule12() Rule12 {
	return Rule12{
		Name:   "rule12",
		MaxAge: time.Duration(12) * time.Second,
		SWR:    time.Duration(12*2) * time.Second,
	}
}

func (r Rule12) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule12) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule12) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule12(a, b Rule12) Rule12 {
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
