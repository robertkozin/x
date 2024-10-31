package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robertkozin/x/internal/httpx"
	"github.com/robertkozin/x/internal/httpx/route"
	"github.com/samber/oops"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
)

var (
	db *pgxpool.Pool
)

//go:generate go run github.com/robertkozin/x/cmd/htmgo

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "run error: %+v", err)
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

	route.Handle(r, "GET", "/comments/{pageId}", getComments)
	route.Handle(r, "POST", "/comments/{pageId}", postComment)

	route.Handle(r, "GET", "/", func(ctx *httpx.Ctx) error {
		return errors.New("404 page not found")
	})

	return r.Mux
}
