package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func ListenAndServe(ctx context.Context, handler http.Handler) {
	var (
		host, port, addr string
		stop, cancel     context.CancelFunc
		ok               bool
	)

	ctx, stop = signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
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
		slog.Info("listening http", "address", server.Addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listening http", "address", server.Addr, "error", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx := context.Background()
	shutdownCtx, cancel = context.WithTimeout(shutdownCtx, 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutting http server down", "address", server.Addr, "error", err)
	}
	slog.Info("done listening http", "address", server.Addr)
}
