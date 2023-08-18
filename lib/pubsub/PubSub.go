package pubsub

import "sync"

// Never copy
type Publisher[T any] struct {
	ev   *dataHolder[T]
	lock *sync.Mutex
	cond *sync.Cond
}

func MakePublisher[T any]() *Publisher[T] {
	var lock = &sync.Mutex{}
	return &Publisher[T]{
		ev: &dataHolder[T]{},
		lock: lock,
		cond: sync.NewCond(lock),
	}
}

func (this *Publisher[T]) Publish(data T) {
	this.lock.Lock()
	this.ev.next = &dataHolder[T]{
		next: nil,
		data: data,
	}
	this.ev = this.ev.next
	this.lock.Unlock()
	this.cond.Broadcast()
}

func (this *Publisher[T]) Subscribe() *Subscription[T] {
	this.lock.Lock()
	defer this.lock.Unlock()
	return &Subscription[T] {
		lock: this.lock,
		cond: this.cond,
		ev: this.ev,
	}
}

type Subscription[T any] struct {
	lock *sync.Mutex
	cond *sync.Cond
	ev *dataHolder[T]
}

func (s *Subscription[T]) Next() T {
	s.lock.Lock()
	defer s.lock.Unlock()
	for s.ev.next == nil {
		s.cond.Wait()
	}
	s.ev = s.ev.next
	return s.ev.data
}

type dataHolder[T any] struct {
	next *dataHolder[T]
	data T
}

