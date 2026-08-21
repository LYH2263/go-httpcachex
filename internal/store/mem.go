package store

import "example.com/httpcachex/internal/entry"

type Mem struct {
	max  int
	data map[string]entry.Entry
	order []string
}

func New(max int) *Mem {
	return &Mem{max: max, data: make(map[string]entry.Entry)}
}

func (m *Mem) Get(k string) (entry.Entry, bool) {
	e, ok := m.data[k]
	if !ok {
		return entry.Entry{}, false
	}
	e.Hits++
	m.data[k] = e
	return entry.CloneEntry(e), true
}

func (m *Mem) Put(k string, e entry.Entry) {
	if _, ok := m.data[k]; !ok {
		if len(m.data) >= m.max && len(m.order) > 0 {
			old := m.order[0]
			m.order = m.order[1:]
			delete(m.data, old)
		}
		m.order = append(m.order, k)
	}
	m.data[k] = entry.CloneEntry(e)
}

func (m *Mem) Delete(k string) {
	delete(m.data, k)
	out := m.order[:0]
	for _, x := range m.order {
		if x != k {
			out = append(out, x)
		}
	}
	m.order = out
}

func (m *Mem) List() []entry.Entry {
	out := make([]entry.Entry, 0, len(m.order))
	for _, k := range m.order {
		if e, ok := m.data[k]; ok {
			out = append(out, entry.CloneEntry(e))
		}
	}
	return out
}

func (m *Mem) Replace(list []entry.Entry) {
	m.data = make(map[string]entry.Entry, len(list))
	m.order = make([]string, 0, len(list))
	for _, e := range list {
		m.data[e.Key] = entry.CloneEntry(e)
		m.order = append(m.order, e.Key)
	}
}

func (m *Mem) Clear() {
	m.data = make(map[string]entry.Entry)
	m.order = nil
}

func (m *Mem) Len() int { return len(m.data) }
