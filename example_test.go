//nolint:goerr113
package shutdown_test

import (
	"context"
	"errors"
	"time"

	"github.com/burik666/shutdown"
)

func ExampleWatch() {
	ctx, ctxCancel := shutdown.Watch(
		// Allows pressing Ctrl+C twice to force shutdown
		shutdown.WithDoubleSignal(),
		// Timeout for graceful shutdown
		shutdown.WithTimeout(30*time.Second))

	defer ctxCancel()

	go func() {
		// Do something
	}()

	<-ctx.Done()

	// shutdown and cleanup
}

func ExampleWithWatcher() {
	ctx, ctxCancel := shutdown.Watch(
		shutdown.WithWatcher(func(ctx context.Context, ch chan<- error) error {
			// If something occurs and you need to shut down the application,
			// send an error to the channel (ch) to trigger the shutdown process.
			ch <- errors.New("shutdown")

			// To shut down the application immediately,
			// send another error to the channel (ch).
			ch <- errors.New("shutdown NOW")

			return nil
		}),
	)

	defer ctxCancel()

	go func() {
		// Do something
	}()

	<-ctx.Done()

	// Shutdown and cleanup
}
