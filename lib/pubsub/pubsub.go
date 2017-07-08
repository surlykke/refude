// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package pubsub

import "sync"

type Event struct {
    next *Event
    data interface{}
}

type endOfStream struct {}


type Publisher struct {
	cond    *sync.Cond
    mutex   *sync.RWMutex
    current *Event
}

type Subscribtion struct {
	p       *Publisher
    current *Event
}

func (p *Publisher) Publish(data interface{}) {
	p.mutex.Lock()
	defer p.cond.Broadcast()
    defer p.mutex.Unlock()
   	 
	p.current.next = &Event{data: data}
    p.current = p.current.next
}

func (p *Publisher) Close() {
	p.Publish(endOfStream{})
}


func MakePublisher() Publisher {
	mutex := & sync.RWMutex{}
    return Publisher {
       current: &Event{},
       mutex: mutex,
       cond: sync.NewCond(mutex.RLocker()),
    }
}

func (p *Publisher) MakeSubscriber() Subscribtion {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return Subscribtion{p, p.current}
}


func (s *Subscribtion) Next() (interface{}, bool) {
	if s.current == nil {
		panic("Calling Get on a closed publisher")
	}

	s.p.mutex.RLock()
	defer s.p.mutex.RUnlock()

	for s.current.next == nil {
		s.p.cond.Wait()
	}

	s.current = s.current.next
	if _,ok := s.current.data.(endOfStream); ok {
		s.current = nil
		return nil, false
	} else {
		return s.current.data, true
	}
}
