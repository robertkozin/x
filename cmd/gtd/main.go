package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robertkozin/x/internal/httpx"
	"github.com/robertkozin/x/internal/httpx/route"
	"github.com/samber/oops"
	"github.com/sashabaranov/go-openai"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
)

var (
	db  *pgxpool.Pool
	oai *openai.Client
)

//go:generate go run github.com/robertkozin/x/cmd/htmgo

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.Error("run", "error", err)
	}
}

func run(ctx context.Context) (err error) {
	t1 := time.Now()

	// database
	db, err = pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return oops.Wrap(err)
	}

	if err = db.Ping(ctx); err != nil {
		return oops.Wrapf(err, "db ping")
	}

	// openai client
	oai = openai.NewClient(os.Getenv("OPENAI_KEY"))
	if _, err = oai.ListModels(ctx); err != nil {
		return oops.Wrapf(err, "oai list models")
	}

	slog.Info("setup", slog.Duration("took", time.Since(t1)))

	// http server
	httpx.ListenAndServe(ctx, routes())

	return nil
}

func routes() http.Handler {
	r := route.New[httpx.Ctx]()

	route.Use(r, func(next func(*httpx.Ctx) error) func(*httpx.Ctx) error {
		return func(ctx *httpx.Ctx) error {
			r := ctx.Request()
			slog.Info("request", "method", r.Method, "path", r.URL)
			return next(ctx)
		}
	})

	route.Use(r, func(next func(*httpx.Ctx) error) func(*httpx.Ctx) error {
		return func(ctx *httpx.Ctx) error {
			if err := next(ctx); err != nil {
				log.Printf("%+v\n", err)
				http.Error(ctx.Response(), fmt.Sprintf("%+v", err), 500)
			}
			return nil
		}
	})

	route.Handle(r, "GET", "/{$}", pageIndex)
	route.HandleHttp(r, "GET", "/voice-note/", http.StripPrefix("/voice-note/", http.FileServerFS(recordingFS{db: db})))

	route.Handle(r, "POST", "/note", noteNew)
	route.Handle(r, "POST", "/note/{id}/delete", noteDelete)
	route.Handle(r, "POST", "/note/{id}/edit", noteEdit)
	route.Handle(r, "POST", "/note/{id}/process", noteProcess)

	return r.Mux
}
