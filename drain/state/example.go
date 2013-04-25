package state

func stop(m *StateMachine, rev int64) {
	// err := drain.Stop()
}

func start(m *StateMachine, rev int64) {
	// start it
	// drain.Start()
	// ...

	go monitor(m, rev+1)
}

func monitor(m *StateMachine, rev int64) {
	var err error
	// err := drain.Wait()
	if err != nil {
		m.SetState(rev, StoppedState, nil)
	} else {
		m.SetState(rev, RetryingState, func() {
			go retry(m, rev+1)
		})
	}
}

func retry(m *StateMachine, rev int64) {
	// this could block.
	if true { // retry.Wait() {
		m.SetState(rev, RunningState, func() {
			start(m, rev)
		})
	} else {
		m.SetState(rev, FatalState, nil)
	}
}
