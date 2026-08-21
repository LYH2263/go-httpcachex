package policy

import "time"

// Rule10 缓存策略片段。
type Rule10 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule10() Rule10 {
	return Rule10{
		Name:   "rule10",
		MaxAge: time.Duration(10) * time.Second,
		SWR:    time.Duration(10*2) * time.Second,
	}
}

func (r Rule10) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule10) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule10) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule10(a, b Rule10) Rule10 {
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
