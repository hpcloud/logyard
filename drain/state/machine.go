package state

import (
	"logyard/util/retry"
	"sync"
	"fmt"
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
			fmt.Printf("state change for '%s' => %s (%d)\n",
				m.process.String(), m.state, m.rev)
		}()
	}
}

func (m *StateMachine) GetState() (State, int64) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if m.IsStopped() {
		panic("stopped")
	}
	return m.state, m.rev
}

func (m *StateMachine) Stop() {
	m.mux.Lock()
	defer m.mux.Unlock()
	close(m.ActionCh)
	m.ActionCh = nil // sentinal to indicate the stopped state.
}

func (m *StateMachine) IsStopped() bool {
	// XXX: ideally we should use locking here, but don't want to
	// introduce a deadlock when called from `SetStateCustom` which
	// also uses locking.
	return m.ActionCh == nil
}

func (m *StateMachine) SetStateCustom(rev int64, fn func() State) int64 {
	m.mux.Lock()
	defer m.mux.Unlock()
	if !m.IsStopped() && rev == m.rev {
		m.state = fn()
		if m.state == nil {
			panic("nil state")
		}
		m.rev += 1
		fmt.Printf("custom state change for '%s' => %s (%d)\n",
			m.process.String(), m.state, m.rev)
		return m.rev
	}
	fmt.Printf("can't set state; rev changed (expected %d, has %d) or stopped (%v)",
		rev, m.rev, m.IsStopped())
	return -1
}

func (m *StateMachine) SetState(rev int64, state State) int64 {
	return m.SetStateCustom(rev, func() State {
		return state
	})
}
