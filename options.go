package shutdown

import (
	"context"
	"os"
	"time"
)

type config struct {
	ctx context.Context

	logFn func(string, ...any)

	signals  []os.Signal
	watchers []func(context.Context, chan<- error) error

	doubleSignal  bool
	forceExitCode int

	timeout         time.Duration
	timeoutExitCode int

	exit func(int)
}

// WithContext allows passing a context to the Watcher.
// By default, it uses context.Background.
func WithContext(ctx context.Context) func(*config) {
	return func(cfg *config) {
		cfg.ctx = ctx
	}
}

// WithLog allows passing a custom log function.
// By default, it uses slog.Error.
func WithLog(logFn func(string, ...any)) func(*config) {
	return func(cfg *config) {
		cfg.logFn = logFn
	}
}

// WithoutLog disables logging.
func WithoutLog() func(*config) {
	return WithLog(func(string, ...any) {})
}

// WithSignals allows passing a list of signals to the shutdown process.
// By default, it listens to os.Interrupt and syscall.SIGTERM.
func WithSignals(signals ...os.Signal) func(*config) {
	return func(cfg *config) {
		cfg.signals = signals
	}
}

// WithoutSignals disables the default signals.
func WithoutSignals() func(*config) {
	return WithSignals()
}

// WithDoubleSignal allows receiving a second signal to force an exit.
func WithDoubleSignal() func(*config) {
	return func(cfg *config) {
		cfg.doubleSignal = true
	}
}

// WithExitFunc allows passing a custom exit function.
// By default, it uses os.Exit.
func WithExitFunc(exit func(int)) func(*config) {
	return func(cfg *config) {
		cfg.exit = exit
	}
}

// WithForceExitCode allows setting a custom exit code for forced exits.
// By default, it uses 1.
func WithForceExitCode(code int) func(*config) {
	return func(cfg *config) {
		cfg.forceExitCode = code
	}
}

// WithTimeout allows setting a timeout for forcing an exit.
func WithTimeout(timeout time.Duration) func(*config) {
	return func(cfg *config) {
		cfg.timeout = timeout
	}
}

// WithTimeoutExitCode allows setting a custom exit code for timeouts.
// By default, it uses 1.
func WithTimeoutExitCode(code int) func(*config) {
	return func(cfg *config) {
		cfg.timeoutExitCode = code
	}
}

// WithWatcher allows custom watchers to be added to the shutdown process.
// Each watcher receives the main context and a channel for reporting errors.
// If a watcher returns an error, the main context is canceled.
// Sending an error to the channel initiates the shutdown process.
// Sending a second error triggers an immediate shutdown.
// The error channel is closed after the watcher completes.
func WithWatcher(watchers ...func(context.Context, chan<- error) error) func(*config) {
	return func(cfg *config) {
		cfg.watchers = append(cfg.watchers, watchers...)
	}
}
