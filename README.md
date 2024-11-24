# Shutdown Manager for Go

[![GitHub license](https://img.shields.io/github/license/burik666/shutdown.svg)](https://github.com/burik666/shutdown/blob/master/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/burik666/shutdown.svg)](https://pkg.go.dev/github.com/burik666/shutdown)

A lightweight library to manage graceful shutdowns in Go applications.
This package simplifies signal handling, ensuring your application can clean up resources properly and exit gracefully under various conditions.

## Features

- Graceful handling of termination signals (`SIGTERM`, `SIGINT`, etc.)
- Timeout-based forced exit to prevent stalling
- Double-signal handling for immediate application termination
- Ability to add custom conditions for terminating the application
- No external dependencies

## Installation

    go get github.com/burik666/shutdown

## Usage

Basic usage

```go
func main() {
    ctx, ctxCancel := shutdown.Watch(
        // Allows pressing Ctrl+C twice to force shutdown
        shutdown.WithDoubleSignal(),
        // Timeout for graceful shutdown
        shutdown.WithTimeout(30*time.Second),
    )

    defer ctxCancel()

    go func() {
        // Do something
    }()

    <-ctx.Done()

    // Shutdown and cleanup
}
```

Custom conditions

```go
func main() {
    ctx, ctxCancel := shutdown.Watch(
        shutdown.WithWatcher(func(ctx context.Context, ch chan<- error) error {
            // If something happens and you need to shut down the application,
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
```

For more examples, see the [examples](examples) directory.

## License

The project is licensed under the [MIT License](LICENSE).
