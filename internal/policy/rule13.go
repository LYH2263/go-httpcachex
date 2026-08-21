package policy

import "time"

// Rule13 缓存策略片段。
type Rule13 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule13() Rule13 {
	return Rule13{
		Name:   "rule13",
		MaxAge: time.Duration(13) * time.Second,
		SWR:    time.Duration(13*2) * time.Second,
	}
}

func (r Rule13) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule13) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule13) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule13(a, b Rule13) Rule13 {
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
