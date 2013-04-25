package state

import (
	"sync"
)

type stateFn func(m *StateMachine, action int, rev int64) stateFn

type StateMachine struct {
	state    stateFn
	rev      int64
	actionCh chan int
	mux      sync.Mutex
}

func (m *StateMachine) Run() {
	m.state = StoppedState
	for action := range m.actionCh {
		func() {
			m.mux.Lock()
			defer m.mux.Unlock()
			m.state = m.state(m, action, m.rev)
			m.rev += 1
		}()
	}
}

func (m *StateMachine) SetState(rev int64, state stateFn, preFn func()) int64 {
	if state == nil {
		panic("nil state")
	}
	m.mux.Lock()
	defer m.mux.Unlock()
	if rev == m.rev {
		if preFn != nil {
			preFn()
		}
		m.state = state
		rev += 1
		return rev
	}
	return -1
}
