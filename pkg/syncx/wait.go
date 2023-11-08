package syncx

import "time"

// Wait when the f() is completed successfully, the channel will be written with struct{}
func Wait(f func()) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		f()
		done <- struct{}{}
		close(done)
	}()
	return done
}
func WaitTimeout(timeout time.Duration, f func()) {
	t := time.NewTimer(timeout)
	wait := Wait(f)

	select {
	case <-wait:
		return
	case <-t.C:
		return
	}
}
