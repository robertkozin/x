package main

import (
	"bytes"
	"context"
	"github.com/aarondl/opt/null"
	"github.com/robertkozin/x/internal/httpx"
	"github.com/samber/oops"
	"io"
	"log"
	"log/slog"
	"path/filepath"
	"strings"
	"time"
)

type Note struct {
	ID         int
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
	RecordedAt time.Time `db:"recorded_at"`
	Filename   string
	Text       string
	Summary    string
	Category   string
	Due        null.Val[time.Time]
}

func noteNew(c *httpx.Ctx) error {
	r := c.Request()

	name, recordedAt, err := parseNoteFilename(r.Header.Get("Filename"))
	if err != nil {
		return oops.Wrap(err)
	}

	slog.Info("new note incoming", "name", name, "size", r.ContentLength)

	if r.ContentLength < 14_000 {
		// 14kb is about 1 second of working audio
		return nil
	}

	b, err := io.ReadAll(r.Body)
	_, err = db.Exec(r.Context(),
		`insert into notes(filename, recorded_at, recording) values ($1, $2, $3) 
on conflict(filename) do update set recording = excluded.recording, recorded_at = excluded.recorded_at`,
		name, recordedAt, b)
	if err != nil {
		err = oops.Wrap(err)

	}

	go func(reader io.Reader, filename string, recordedAt time.Time) {
		err := processNote(context.Background(), reader, filename, recordedAt)
		if err != nil {
			log.Printf("failed to process note: %+v\n", err)
		}
	}(bytes.NewReader(b), name, recordedAt)

	return err
}

func processNote(ctx context.Context, reader io.Reader, filename string, recordedAt time.Time) error {
	errb := oops.With("note", filename)

	text, err := transcribe(ctx, reader, filename)
	if err != nil {
		return errb.Wrap(err)
	}
	slog.Info("transcribe", "file", filename, "text", text)

	_, err = db.Exec(ctx, `update notes set text = $1 where filename = $2`, text, filename)
	if err != nil {
		return errb.Wrap(err)
	}

	ret, err := extract(ctx, text, recordedAt)
	if err != nil {
		return errb.Wrap(err)
	}
	slog.Info("extract", "file", filename, "summary", ret.Summary, "cat", ret.Category, "due_date", ret.DueDate, "due_time", ret.DueTime)

	_, err = db.Exec(ctx,
		`update notes set summary=$1, category=$2, due=$3 where filename=$4`,
		ret.Summary, ret.Category, ret.ParsedDueDate, filename)

	return errb.Wrap(err)
}

func parseNoteFilename(filename string) (name string, recordedAt time.Time, err error) {
	errb := oops.With("filename", filename)
	name = filepath.Base(filename) // Tasker sends the entire file path

	rawTime, hasMp4 := strings.CutSuffix(name, ".mp4")
	if !hasMp4 {
		err = errb.Errorf("missing .mp4 suffix")
		return
	}

	recordedAt, err = time.Parse("2006-01-02_15-04-05", rawTime)
	err = errb.Wrap(err)
	return
}

func noteDelete(c *httpx.Ctx) error {
	r := c.Request()
	id := r.PathValue("id")

	_, err := db.Exec(r.Context(), `delete from notes where id = $1`, id)
	err = oops.Wrap(err)

	return err
}

func noteEdit(c *httpx.Ctx) error {
	return nil
}

func noteProcess(c *httpx.Ctx) error {
	return nil
}
