package metrics

import "sync"

// Series2 命中率序列。
type Series2 struct {
	mu   sync.Mutex
	hits int64
	miss int64
	reval int64
}

func NewSeries2() *Series2 { return &Series2{} }

func (s *Series2) Hit()   { s.add(&s.hits) }
func (s *Series2) Miss()  { s.add(&s.miss) }
func (s *Series2) Reval() { s.add(&s.reval) }

func (s *Series2) add(p *int64) {
	s.mu.Lock()
	*p++
	s.mu.Unlock()
}

func (s *Series2) Rate() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := s.hits + s.miss
	if t == 0 {
		return 0
	}
	return float64(s.hits) / float64(t)
}

func (s *Series2) Snapshot() map[string]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]int64{"hits": s.hits, "miss": s.miss, "reval": s.reval}
}
