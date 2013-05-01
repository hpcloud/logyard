package state

import (
	"fmt"
	"github.com/ActiveState/log"
	"logyard/util/retry"
	"sync"
)

type StateMachine struct {
	running bool
	title   string
	process Process
	retryer retry.Retryer
	state   State
	rev     int64
	mux     sync.Mutex
}

func NewStateMachine(title string, process Process, retryer retry.Retryer) *StateMachine {
	m := &StateMachine{}
	m.title = title
	m.process = process
	m.retryer = retryer
	m.state = Stopped{m}
	m.rev = 1
	m.running = true
	return m
}

func (m *StateMachine) Log(msg string, v ...interface{}) {
	log.Infof(m.process.Logf("[STM] "+msg, v...))
}

func (m *StateMachine) SendAction(action int) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if !m.running {
		return fmt.Errorf("StateMachine is stopped")
	}
	oldState := m.state
	m.Log("Received action %s on state %s (%d)\n",
		getActionString(action), oldState, m.rev)
	m.state = m.state.Transition(action, m.rev)
	m.rev += 1
	m.Log("State change: %s (%d) => %s (%d)\n",
		oldState, m.rev-1, m.state, m.rev)
	return nil
}

func (m *StateMachine) GetState() State {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.state
}

func (m *StateMachine) Stop() {
	m.Log("Stopping STM...")
	m.mux.Lock()
	defer m.mux.Unlock()

	m.running = false

	// reset fields to prevent (buggy) future use
	m.state = nil
	m.rev = -10

	m.Log("Stopped STM.")
}

func (m *StateMachine) setStateCustom(rev int64, fn func() State) int64 {
	m.mux.Lock()
	defer m.mux.Unlock()
	if m.running && rev == m.rev {
		oldState := m.state
		m.state = fn()
		if m.state == nil {
			panic("nil state")
		}
		m.rev += 1
		m.Log("State change (custom): %s (%d) => %s (%d)\n",
			oldState, m.rev-1, m.state, m.rev)
		return m.rev
	} else {
		var msg string
		if m.running {
			msg = fmt.Sprintf("rev changed (expected %d, have %d)",
				rev, m.rev)
		} else {
			msg = "STM is stopped"
		}
		m.Log("Skipping state change; %s\n", msg)
		return -1
	}
	panic("unreachable")
}

func (m *StateMachine) setState(rev int64, state State) int64 {
	return m.setStateCustom(rev, func() State {
		return state
	})
}

func (s *StateMachine) stop(rev int64) State {
	s.Log("Stopping %s", s.title)
	err := s.process.Stop()
	if err != nil {
		// Error reporting?
		return Fatal{err, s}
	}
	return Stopped{s}
}

func (s *StateMachine) start(rev int64) State {
	s.Log("Starting %s", s.title)
	err := s.process.Start()
	if err != nil {
		return Fatal{err, s}
	} else {
		rev = rev + 1 // account for settig of RunningState
		go s.monitor(rev)
		return Running{s}
	}
	panic("unreachable")
}

func (s *StateMachine) monitor(rev int64) {
	err := s.process.Wait()

	s.Log("%s exited -- %v", s.title, err)
	if err == nil {
		// If a process exited cleanly (no errors), then just mark it
		// as STOPPED without retrying.
		s.setState(rev, Stopped{s}) // rev confict here is normal.
	} else {
		s.setStateCustom(rev, func() State {
			rev = rev + 1 // account for setting of RetryingState
			go s.doretry(rev, err)
			return Retrying{err, s}
		})
	}
}

func (s *StateMachine) doretry(rev int64, err error) {
	retryMsg := fmt.Sprintf(s.process.Logf(
		"[STM] %s exited abruptly -- %v", s.title, err))
	// retryer.Wait generally blocks on time.Sleep.
	if s.retryer.Wait(retryMsg) {
		s.setStateCustom(rev, func() State {
			return s.start(rev)
		})
	} else {
		err := fmt.Errorf("Retried too long; last error: %v", err)
		s.Log("%v -- marking as FATAL", err)
		s.setState(rev, Fatal{err, s})
	}
}
