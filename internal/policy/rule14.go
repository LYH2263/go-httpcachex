package policy

import "time"

// Rule14 缓存策略片段。
type Rule14 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule14() Rule14 {
	return Rule14{
		Name:   "rule14",
		MaxAge: time.Duration(14) * time.Second,
		SWR:    time.Duration(14*2) * time.Second,
	}
}

func (r Rule14) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule14) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule14) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule14(a, b Rule14) Rule14 {
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
