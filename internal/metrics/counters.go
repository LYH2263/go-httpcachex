package metrics

import "sync"

type Registry struct {
	mu    sync.Mutex
	hit   int64
	miss  int64
	stale int64
	reval int64
}

func New() *Registry { return &Registry{} }

func (r *Registry) IncHit()        { r.add(&r.hit) }
func (r *Registry) IncMiss()       { r.add(&r.miss) }
func (r *Registry) IncStale()      { r.add(&r.stale) }
func (r *Registry) IncRevalidate() { r.add(&r.reval) }

func (r *Registry) add(p *int64) {
	r.mu.Lock()
	*p++
	r.mu.Unlock()
}

func (r *Registry) Snapshot() map[string]int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return map[string]int64{"hit": r.hit, "miss": r.miss, "stale": r.stale, "revalidate": r.reval}
}
