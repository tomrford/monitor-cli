package relay

import (
	"sync"
	"time"
)

type Dedupe struct {
	mu   sync.Mutex
	seen map[string]time.Time
	ttl  time.Duration
}

func NewDedupe(ttl time.Duration) *Dedupe {
	return &Dedupe{
		seen: map[string]time.Time{},
		ttl:  ttl,
	}
}

func (d *Dedupe) Seen(id string, now time.Time) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if exp, ok := d.seen[id]; ok && now.Before(exp) {
		return true
	}
	d.seen[id] = now.Add(d.ttl)
	return false
}

func (d *Dedupe) Prune(now time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for k, exp := range d.seen {
		if !now.Before(exp) {
			delete(d.seen, k)
		}
	}
}
