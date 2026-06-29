package pool

import "sync"

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	mu    sync.Mutex
	items []T
	new   func() T
}

func New[T Resettable](new func() T) *Pool[T] {
	return &Pool[T]{
		items: make([]T, 0),
		new:   new,
	}
}

func (p *Pool[T]) Get() T {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.items) > 0 {
		item := p.items[len(p.items)-1]
		p.items = p.items[:len(p.items)-1]
		return item
	}

	if p.new != nil {
		return p.new()
	}

	var zero T
	return zero
}

func (p *Pool[T]) Put(item T) {
	item.Reset()
	p.mu.Lock()
	defer p.mu.Unlock()
	p.items = append(p.items, item)
}
