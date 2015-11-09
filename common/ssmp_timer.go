package common

import (
	"time"
)

const (
	FOREVER = 10000 * time.Hour
)

type PTimer struct {
	tout  time.Duration
	timer *time.Timer
}

func NewPTimer(timeout time.Duration) *PTimer {
	ptimer := &PTimer{
		tout:  timeout,
		timer: time.NewTimer(FOREVER),
	}

	return ptimer
}

func (pt *PTimer) TimerOn() {
	pt.timer.Stop()
	pt.timer = time.NewTimer(pt.tout)
}

func (pt *PTimer) TimerOff() {
	pt.timer.Stop()
	pt.timer = time.NewTimer(FOREVER)
}

func (pt *PTimer) C() <-chan time.Time {
	return pt.timer.C
}
