package policy

import "time"

// Rule9 缓存策略片段。
type Rule9 struct {
	Name     string
	MaxAge   time.Duration
	SWR      time.Duration
	NoStore  bool
	MustRevalidate bool
}

func DefaultRule9() Rule9 {
	return Rule9{
		Name:   "rule9",
		MaxAge: time.Duration(9) * time.Second,
		SWR:    time.Duration(9*2) * time.Second,
	}
}

func (r Rule9) FreshUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge)
}

func (r Rule9) StaleUntil(now time.Time) time.Time {
	return now.Add(r.MaxAge + r.SWR)
}

func (r Rule9) CanServeStale(now, staleAt time.Time) bool {
	if r.NoStore || r.MustRevalidate {
		return false
	}
	return now.Before(staleAt.Add(r.SWR))
}

func MergeRule9(a, b Rule9) Rule9 {
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
