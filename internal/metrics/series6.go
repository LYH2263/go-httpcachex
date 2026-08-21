package metrics

import "sync"

// Series6 命中率序列。
type Series6 struct {
	mu   sync.Mutex
	hits int64
	miss int64
	reval int64
}

func NewSeries6() *Series6 { return &Series6{} }

func (s *Series6) Hit()   { s.add(&s.hits) }
func (s *Series6) Miss()  { s.add(&s.miss) }
func (s *Series6) Reval() { s.add(&s.reval) }

func (s *Series6) add(p *int64) {
	s.mu.Lock()
	*p++
	s.mu.Unlock()
}

func (s *Series6) Rate() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := s.hits + s.miss
	if t == 0 {
		return 0
	}
	return float64(s.hits) / float64(t)
}

func (s *Series6) Snapshot() map[string]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]int64{"hits": s.hits, "miss": s.miss, "reval": s.reval}
}
