package state

import (
	"fmt"
	"logyard/util/retry"
	"sync"
)

type StateMachine struct {
	ActionCh chan int
	process  Process
	retryer  retry.Retryer
	state    State
	rev      int64
	mux      sync.Mutex
}

func NewStateMachine(process Process, retryer retry.Retryer) *StateMachine {
	m := &StateMachine{}
	m.process = process
	m.retryer = retryer
	m.state = Stopped{m}
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
			m.state = m.state.Transition(action, m.rev)
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

func (s *StateMachine) stop(rev int64) State {
	err := s.process.Stop()
	if err != nil {
		return Fatal{s}
	}
	return Stopped{s}
}

func (s *StateMachine) start(rev int64) State {
	// start it
	err := s.process.Start()
	if err != nil {
		return Fatal{s}
	} else {
		rev = rev + 1 // account for settig of RunningState
		go s.monitor(rev)
		return Running{s}
	}
	panic("unreachable")
}

func (s *StateMachine) monitor(rev int64) {
	err := s.process.Wait()
	if err == nil {
		s.SetState(rev, Stopped{s})
	} else {
		s.SetStateCustom(rev, func() State {
			rev = rev + 1 // account for setting of RetryingState
			go s.doretry(rev, err)
			return Retrying{s}
		})
	}
}

func (s *StateMachine) doretry(rev int64, err error) {
	// This could block.
	if s.retryer.Wait(
		fmt.Sprintf("[drain:???] Drain exited abruptly -- %v", err)) {
		// TODO: move 'drain' specific message (above) out of the
		// state package.
		s.SetStateCustom(rev, func() State {
			fmt.Println("starting ???")
			return s.start(rev)
		})
	} else {
		s.SetState(rev, Fatal{s})
	}
}
