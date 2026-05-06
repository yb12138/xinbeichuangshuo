package server

import "sync/atomic"

type roomActorCall struct {
	fn    func() error
	errCh chan error
	onErr func(error)
}

func (r *Room) executeViaActor(fn func() error) error {
	if fn == nil {
		return nil
	}
	if !r.actorLoopStarted.Load() {
		return fn()
	}

	call := roomActorCall{
		fn:    fn,
		errCh: make(chan error, 1),
	}
	r.actorInbox <- call
	return <-call.errCh
}

func (r *Room) enqueueViaActor(fn func() error, onErr func(error)) {
	if fn == nil {
		return
	}
	if !r.actorLoopStarted.Load() {
		if err := fn(); err != nil && onErr != nil {
			onErr(err)
		}
		return
	}

	r.actorInbox <- roomActorCall{
		fn:    fn,
		onErr: onErr,
	}
}

func (r *Room) handleActorCall(call roomActorCall) {
	if call.fn == nil {
		if call.errCh != nil {
			call.errCh <- nil
			close(call.errCh)
		}
		return
	}

	err := call.fn()
	if call.onErr != nil && err != nil {
		call.onErr(err)
	}
	if call.errCh != nil {
		call.errCh <- err
		close(call.errCh)
	}
}

type actorLoopFlag struct {
	atomic.Bool
}
