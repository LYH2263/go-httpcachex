package metrics

import "sync"

// Series1 命中率序列。
type Series1 struct {
	mu   sync.Mutex
	hits int64
	miss int64
	reval int64
}

func NewSeries1() *Series1 { return &Series1{} }

func (s *Series1) Hit()   { s.add(&s.hits) }
func (s *Series1) Miss()  { s.add(&s.miss) }
func (s *Series1) Reval() { s.add(&s.reval) }

func (s *Series1) add(p *int64) {
	s.mu.Lock()
	*p++
	s.mu.Unlock()
}

func (s *Series1) Rate() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := s.hits + s.miss
	if t == 0 {
		return 0
	}
	return float64(s.hits) / float64(t)
}

func (s *Series1) Snapshot() map[string]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]int64{"hits": s.hits, "miss": s.miss, "reval": s.reval}
}
