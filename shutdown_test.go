//nolint:goerr113
package shutdown_test

import (
	"context"
	"errors"
	"syscall"
	"testing"
	"time"

	"github.com/burik666/shutdown"
)

const settleTime = 250 * time.Millisecond

func waitDone(t *testing.T, ctx context.Context) {
	t.Helper()

	select {
	case <-ctx.Done():
		return
	case <-time.After(settleTime):
		t.Fatal("timeout reached")
	}

	t.Fatal("unexpected")
}

func waitDoneWithExit(t *testing.T, ctx context.Context, exitCh <-chan int) {
	t.Helper()

	select {
	case <-ctx.Done():
		select {
		case code := <-exitCh:
			if code != 2 {
				t.Fatalf("unexpected exit code: %d", code)
			}

			return

		case <-time.After(settleTime):
			t.Fatal("shutdown timeout reached")
		}

	case <-time.After(settleTime):
		t.Fatal("timeout reached")
	}

	t.Fatal("unexpected")
}

// func TestMain(m *testing.M) {
//     goleak.VerifyTestMain(m)
// }

func TestShutdown(t *testing.T) {
	ctx, _ := shutdown.Watch(
		shutdown.WithoutLog(),
		shutdown.WithSignals(syscall.SIGUSR1),
	)

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1); err != nil {
		t.Fatal(err)
	}

	waitDone(t, ctx)
}

func TestDoubleSignal(t *testing.T) {
	exitCh := make(chan int, 1)
	defer close(exitCh)

	ctx, _ := shutdown.Watch(
		shutdown.WithoutLog(),
		shutdown.WithSignals(syscall.SIGUSR1),
		shutdown.WithExitFunc(func(code int) {
			exitCh <- code
		}),
		shutdown.WithDoubleSignal(),
		shutdown.WithForceExitCode(2),
		shutdown.WithTimeout(settleTime),
	)

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1); err != nil {
		t.Fatal(err)
	}

	time.Sleep(settleTime / 10)

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1); err != nil {
		t.Fatal(err)
	}

	waitDoneWithExit(t, ctx, exitCh)
}

func TestTimeout(t *testing.T) {
	exitCh := make(chan int, 1)
	defer close(exitCh)

	ctx, _ := shutdown.Watch(
		shutdown.WithoutLog(),
		shutdown.WithSignals(syscall.SIGUSR1),
		shutdown.WithExitFunc(func(code int) {
			exitCh <- code
		}),
		shutdown.WithDoubleSignal(),
		shutdown.WithTimeout(settleTime/4),
		shutdown.WithTimeoutExitCode(2),
	)

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1); err != nil {
		t.Fatal(err)
	}

	waitDoneWithExit(t, ctx, exitCh)
}

func TestCtx(t *testing.T) {
	type ctxKey struct{}

	pctx := context.WithValue(context.Background(), ctxKey{}, "somevalue")

	ctx, cancel := shutdown.Watch(
		shutdown.WithContext(pctx),
	)

	if ctx.Value(ctxKey{}) != "somevalue" {
		t.Fatal("unexpected context value")
	}

	cancel()

	waitDone(t, ctx)
}

func TestWatcherError(t *testing.T) {
	ctx, _ := shutdown.Watch(
		shutdown.WithoutLog(),
		shutdown.WithoutSignals(),
		shutdown.WithWatcher(func(_ context.Context, _ chan<- error) error {
			return errors.New("some error")
		}),
	)

	waitDone(t, ctx)
}

func TestWatcher(t *testing.T) {
	ctx, _ := shutdown.Watch(
		shutdown.WithoutLog(),
		shutdown.WithoutSignals(),
		shutdown.WithWatcher(func(ctx context.Context, ch chan<- error) error {
			ch <- errors.New("shutdown")

			return nil
		}),
	)

	waitDone(t, ctx)
}

func TestNopWatcher(t *testing.T) {
	ctx, _ := shutdown.Watch(
		shutdown.WithoutLog(),
		shutdown.WithoutSignals(),
		shutdown.WithWatcher(func(ctx context.Context, ch chan<- error) error {
			return nil
		}),
	)

	select {
	case <-ctx.Done():
		t.Fatal("unexpected")
	case <-time.After(settleTime):
		return
	}

	t.Fatal("unexpected")
}

func TestWatcherForce(t *testing.T) {
	exitCh := make(chan int, 1)
	defer close(exitCh)

	ctx, _ := shutdown.Watch(
		shutdown.WithoutLog(),
		shutdown.WithExitFunc(func(code int) {
			exitCh <- code
		}),
		shutdown.WithForceExitCode(2),
		shutdown.WithoutSignals(),
		shutdown.WithWatcher(func(ctx context.Context, ch chan<- error) error {
			ch <- errors.New("shutdown")
			ch <- errors.New("shutdown NOW")

			return nil
		}),
	)

	waitDoneWithExit(t, ctx, exitCh)
}

func TestWatcherTimeout(t *testing.T) {
	exitCh := make(chan int, 1)
	defer close(exitCh)

	ctx, _ := shutdown.Watch(
		shutdown.WithoutLog(),
		shutdown.WithExitFunc(func(code int) {
			exitCh <- code
		}),
		shutdown.WithoutSignals(),
		shutdown.WithTimeout(settleTime/4),
		shutdown.WithTimeoutExitCode(2),
		shutdown.WithWatcher(func(ctx context.Context, ch chan<- error) error {
			ch <- errors.New("shutdown")

			return nil
		}),
	)

	waitDoneWithExit(t, ctx, exitCh)
}

func TestDumbWatcher(t *testing.T) {
	_, _ = shutdown.Watch(
		shutdown.WithoutSignals(),
	)
}
