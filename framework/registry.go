package framework

import "sync"

type Registry struct {
	mu    sync.RWMutex
	nodes map[string]Drawable
}

func NewRegistry() *Registry {
	return &Registry{nodes: make(map[string]Drawable)}
}

func (r *Registry) Register(id string, d Drawable) (collided bool) {
	if r == nil || id == "" {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, exists := r.nodes[id]
	r.nodes[id] = d
	return exists && existing != d
}

func (r *Registry) Unregister(id string) {
	if r == nil || id == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.nodes, id)
}

func (r *Registry) Find(id string) (Drawable, bool) {
	if r == nil {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.nodes[id]
	return d, ok
}

func FindAs[T Drawable](r *Registry, id string) (T, bool) {
	var zero T
	d, ok := r.Find(id)
	if !ok {
		return zero, false
	}
	t, ok := d.(T)
	if !ok {
		return zero, false
	}
	return t, true
}
