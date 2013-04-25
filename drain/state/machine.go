package state

import (
	"logyard/util/retry"
	"sync"
)

type State func(m *StateMachine, action int, rev int64) State

type StateMachine struct {
	ActionCh chan int
	process  Process
	retryer  retry.Retryer
	state    State
	rev      int64
	mux      sync.Mutex
}

func New(process Process, retryer retry.Retryer) *StateMachine {
	m := &StateMachine{}
	m.process = process
	m.retryer = retryer
	m.state = StoppedState
	m.rev = 1
	m.ActionCh = make(chan int)
	go m.Run()
	return m
}

func (m *StateMachine) Run() {
	for action := range m.ActionCh {
		func() {
			m.mux.Lock()
			defer m.mux.Unlock()
			m.state = m.state(m, action, m.rev)
			m.rev += 1
		}()
	}
}

func (m *StateMachine) GetState() (State, int64) {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.state, m.rev
}

func (m *StateMachine) Stop() {
	// TODO: use tomb.
	close(m.ActionCh)
}

func (m *StateMachine) SetStateCustom(rev int64, fn func() State) int64 {
	m.mux.Lock()
	defer m.mux.Unlock()
	if rev == m.rev {
		m.state = fn()
		if m.state == nil {
			panic("nil state")
		}
		rev += 1
		return rev
	}
	return -1
}

func (m *StateMachine) SetState(rev int64, state State) int64 {
	return m.SetStateCustom(rev, func() State {
		return state
	})
}
