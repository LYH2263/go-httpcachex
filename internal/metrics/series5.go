package metrics

import "sync"

// Series5 命中率序列。
type Series5 struct {
	mu   sync.Mutex
	hits int64
	miss int64
	reval int64
}

func NewSeries5() *Series5 { return &Series5{} }

func (s *Series5) Hit()   { s.add(&s.hits) }
func (s *Series5) Miss()  { s.add(&s.miss) }
func (s *Series5) Reval() { s.add(&s.reval) }

func (s *Series5) add(p *int64) {
	s.mu.Lock()
	*p++
	s.mu.Unlock()
}

func (s *Series5) Rate() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := s.hits + s.miss
	if t == 0 {
		return 0
	}
	return float64(s.hits) / float64(t)
}

func (s *Series5) Snapshot() map[string]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]int64{"hits": s.hits, "miss": s.miss, "reval": s.reval}
}
