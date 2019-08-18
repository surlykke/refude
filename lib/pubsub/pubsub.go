package pubsub

import (
	"sync"
)

type event struct {
	path string
	next *event
}

type Subscription struct {
	lock    *sync.RWMutex
	cond    *sync.Cond
	current *event
}

type PubSub struct {
	event *event
	lock  *sync.RWMutex
	cond  *sync.Cond
}

func MakePubSub() *PubSub {
	var ps = &PubSub{}
	ps.event = &event{path: "", next: nil}
	ps.lock = &sync.RWMutex{}
	ps.cond = &sync.Cond{L: ps.lock}
	return ps
}

func (ps *PubSub) Publish(path string) {
	ps.lock.Lock()
	ps.event.next = &event{path, nil}
	ps.event = ps.event.next
	ps.lock.Unlock()
	ps.cond.Broadcast()
}

func (ps *PubSub) Subscribe() *Subscription {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	return &Subscription{ps.lock, ps.cond, ps.event}
}

func (s *Subscription) WaitFor(path string) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for {
		for s.current.next == nil {
			s.cond.Wait()
		}

		s.current = s.current.next
		if s.current.path == path {
			return
		}
	}
}
