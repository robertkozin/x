package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func ListenAndServe(ctx context.Context, handler http.Handler) {
	var (
		host, port, addr string
		stop, cancel     context.CancelFunc
		ok               bool
	)

	ctx, stop = signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	if host, ok = os.LookupEnv("HOST"); !ok {
		host = "0.0.0.0"
	}
	if port, ok = os.LookupEnv("PORT"); !ok {
		port = "8080"
	}
	addr = net.JoinHostPort(host, port)

	server := &http.Server{Addr: addr, Handler: handler}

	go func() {
		slog.Info("Listening", "address", server.Addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Error listening", "address", server.Addr, "error", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx := context.Background()
	shutdownCtx, cancel = context.WithTimeout(shutdownCtx, 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Error shutting down server", "address", server.Addr, "error", err)
	}
	slog.Info("Done listening", "address", server.Addr)
}
