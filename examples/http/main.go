package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/burik666/shutdown"
)

func main() {
	ctx, ctxCancel := shutdown.Watch(
		// Allows pressing Ctrl+C twice to force shutdown
		shutdown.WithDoubleSignal(),
		// Timeout for graceful shutdown
		shutdown.WithTimeout(30*time.Second))

	defer ctxCancel()

	// HTTP server
	srv := http.Server{
		Addr: "localhost:8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("Hello, World!"))
		}),
	}

	go func() {
		log.Printf("http server started on %s", srv.Addr)

		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Printf("http server error: %s", err)

			// Shutdown on error
			ctxCancel()
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	log.Printf("http server stopping...")

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("http server shutdown error: %s", err)
	}

	log.Printf("http server stopped, bye!")
}
