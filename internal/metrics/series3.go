package metrics

import "sync"

// Series3 命中率序列。
type Series3 struct {
	mu   sync.Mutex
	hits int64
	miss int64
	reval int64
}

func NewSeries3() *Series3 { return &Series3{} }

func (s *Series3) Hit()   { s.add(&s.hits) }
func (s *Series3) Miss()  { s.add(&s.miss) }
func (s *Series3) Reval() { s.add(&s.reval) }

func (s *Series3) add(p *int64) {
	s.mu.Lock()
	*p++
	s.mu.Unlock()
}

func (s *Series3) Rate() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := s.hits + s.miss
	if t == 0 {
		return 0
	}
	return float64(s.hits) / float64(t)
}

func (s *Series3) Snapshot() map[string]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]int64{"hits": s.hits, "miss": s.miss, "reval": s.reval}
}
