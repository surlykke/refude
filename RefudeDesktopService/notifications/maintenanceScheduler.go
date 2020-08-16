package notifications

import (
	"time"
)

/**
 * Simple wrapper around time.Timer, to allow scheduling of pings (can be received through the ping channel)
 *
 */
type maintenanceScheduler struct {
	ping     chan struct{}
	timer    *time.Timer
	schedule chan time.Time
	drained  bool
}

func makeScheduler() *maintenanceScheduler {
	return &maintenanceScheduler{
		ping:     make(chan struct{}),
		timer:    time.NewTimer(3600 * time.Second),
		schedule: make(chan time.Time),
		drained:  false,
	}
}

func (ms maintenanceScheduler) run() {
	for {
		select {
		case t := <-ms.schedule:
			if !ms.timer.Stop() && !ms.drained {
				<-ms.timer.C
			}
			ms.timer.Reset(t.Sub(time.Now()))
			ms.drained = false
		case <-ms.timer.C:
			ms.drained = true
			go func() { ms.ping <- struct{}{} }()
		}
	}
}
