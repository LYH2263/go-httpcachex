package policy

import "time"

// Rule7 缓存策略片段。
type Rule7 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule7() Rule7 {
	return Rule7{
		Name:   "rule7",
		MaxAge: time.Duration(7) * time.Second,
		SWR:    time.Duration(7*2) * time.Second,
	}
}

func (r Rule7) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule7) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule7) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule7(a, b Rule7) Rule7 {
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
