package shutdown

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Watch starts a shutdown watcher.
func Watch(opts ...func(*config)) (context.Context, context.CancelFunc) {
	cfg := config{
		ctx:   context.Background(),
		logFn: slog.Error,

		signals:         []os.Signal{os.Interrupt, syscall.SIGTERM},
		exit:            os.Exit,
		forceExitCode:   1,
		timeoutExitCode: 1,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	ctx, ctxCancel := context.WithCancelCause(cfg.ctx)

	if len(cfg.signals) > 0 {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, cfg.signals...)

		WithWatcher(func(ctx context.Context, ch chan<- error) error {
			defer close(sigCh)
			defer signal.Stop(sigCh)

			select {
			case <-ctx.Done():
				return nil
			case sig := <-sigCh:
				ch <- fmt.Errorf("signal %q received", sig.String()) //nolint:goerr113

				if cfg.doubleSignal {
					shutdownCtx, shutdownCtxCancel := context.WithCancel(context.Background())

					if cfg.timeout > 0 {
						shutdownCtx, shutdownCtxCancel = context.WithTimeout(shutdownCtx, cfg.timeout)
					}

					defer shutdownCtxCancel()

					select {
					case <-shutdownCtx.Done():
						return nil
					case sig = <-sigCh:
						ch <- fmt.Errorf("second %q signal received", sig.String()) //nolint:goerr113
					}
				}

				return nil
			}
		})(&cfg)
	}

	var wg sync.WaitGroup

	wg.Add(len(cfg.watchers))

	for _, w := range cfg.watchers {
		go func(watcher func(context.Context, chan<- error) error) {
			ch := make(chan error, 1)

			go func() {
				defer close(ch)

				wg.Done()

				if err := watcher(ctx, ch); err != nil {
					cfg.logFn(fmt.Sprintf("watcher error: %s", err))

					ctxCancel(err)
				}
			}()

			select {
			case <-ctx.Done():
				return
			case err, ok := <-ch:
				if !ok {
					return
				}

				if err != nil {
					cfg.logFn(err.Error())
				}

				ctxCancel(err)

				shutdownCtx, shutdownCtxCancel := context.WithCancel(context.Background())

				if cfg.timeout > 0 {
					shutdownCtx, shutdownCtxCancel = context.WithTimeout(shutdownCtx, cfg.timeout)
				}

				defer shutdownCtxCancel()

				for {
					select {
					case <-shutdownCtx.Done():
						cfg.logFn("shutdown timeout reached, force exit")

						cfg.exit(cfg.timeoutExitCode)

						return

					case err, ok := <-ch:
						if !ok {
							if cfg.timeout > 0 {
								continue
							}

							return
						}

						if err != nil {
							cfg.logFn(err.Error())
						}

						cfg.exit(cfg.forceExitCode)

						return
					}
				}
			}
		}(w)
	}

	wg.Wait()

	return ctx, func() { ctxCancel(nil) }
}
