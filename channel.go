package dvotcWS

import (
	"sync"
)

type FIFOQueue[T any] struct {
	c     chan *T
	mutex sync.Mutex
}

func NewFIFOQueue[T LevelData]() *FIFOQueue[T] {
	return &FIFOQueue[T]{c: make(chan *T, 1)}
}

func (q *FIFOQueue[T]) enqueue(item *T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	for len(q.c) > 0 {
		<-q.c
	}
	q.c <- item
}

func (q *FIFOQueue[T]) Dequeue() (*T, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	select {
	case item := <-q.c:
		return item, true
	default:
		return nil, false
	}
}
