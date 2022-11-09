package mybench

import (
	"container/ring"
	"sync"
)

// A terrible implementation of a ring, based on the Golang ring which is not
// thread-safe nor offers a nice API.
//
// I can't believe there are no simple ring buffer data structure in Golang,
// with generics.
type Ring[T any] struct {
	mut      *sync.Mutex
	capacity int
	ring     *ring.Ring
}

func NewRing[T any](capacity int) *Ring[T] {
	return &Ring[T]{
		mut:      &sync.Mutex{},
		capacity: capacity,
		ring:     ring.New(capacity),
	}
}

func (r *Ring[T]) Push(data T) {
	r.mut.Lock()
	defer r.mut.Unlock()

	r.ring = r.ring.Next()
	r.ring.Value = data
}

func (r *Ring[T]) ReadAllOrdered() []T {
	arr := make([]T, 0, r.capacity)

	r.mut.Lock()
	defer r.mut.Unlock()

	earliest := r.ring

	for earliest.Prev() != nil && earliest.Prev() != r.ring && earliest.Prev().Value != nil {
		earliest = earliest.Prev()
	}

	for earliest != r.ring {
		arr = append(arr, earliest.Value.(T))
		earliest = earliest.Next()
	}

	if earliest.Value == nil {
		return arr
	}

	arr = append(arr, earliest.Value.(T))

	return arr
}
