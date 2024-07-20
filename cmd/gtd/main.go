package main

import (
	"cmp"
	"context"
	"fmt"
	"github.com/robertkozin/x/htmgo"
	"github.com/robertkozin/x/internal/httpx"
	"github.com/robertkozin/x/internal/httpx/route"
	"github.com/robertkozin/x/internal/jsonbackup"
	"github.com/sashabaranov/go-openai"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	oai   = openai.NewClient(os.Getenv("OPENAI_KEY"))
	notes = make(map[int]*Note)
	mu    sync.RWMutex

	voiceNotesDir = "./voice-note"
)

type Note struct {
	ID         int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	CapturedAt time.Time
	Filename   string
	Text       string
	Summary    string
	Topic      string
	Due        time.Time
}

//go:generate go run github.com/robertkozin/x/cmd/htmgo

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatalln(err)
	}
}

func run(ctx context.Context) (err error) {
	if err = os.MkdirAll(voiceNotesDir, os.ModePerm); err != nil {
		return err
	}

	defer jsonbackup.Must(&notes, "./notes.json")()

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
				http.Error(ctx.Response(), fmt.Sprintf("%+v", err), 500)
			}
			return nil
		}
	})

	route.Handle(r, "GET", "/{$}", index)
	route.HandleHttp(r, "GET", "/voice-note/", http.FileServer(http.Dir(".")))
	route.Handle(r, "POST", "/note", postNote)
	route.Handle(r, "POST", "/note/{id}/delete", deleteNote)
	route.Handle(r, "POST", "/note/{id}/edit", editNote)
	route.Handle(r, "POST", "/note/{id}/process", processNote)

	httpx.ListenAndServe(ctx, r.Mux)

	return nil
}

func index(c *httpx.Ctx) (err error) {
	mu.RLock()
	defer mu.RUnlock()

	notes := sortMap(notes, func(a *Note, b *Note) int {
		return b.ID - a.ID
	})

	return htmgo.Render(context.Background(), c.Response(), Index{notes: notes})
}

func (i Index) NoteTitle(note *Note, lastNote *Note) string {
	t := note.Summary
	if t == "" {
		t = note.Text
	}
	if t == "" {
		t = `(none)`
	} else if len(t) > 150 {
		t = t[:150]
		t = t + "..."
	}

	if note.Topic != "" {
		t = fmt.Sprintf("[%s] %s", note.Topic, t)
	}

	if !note.Due.IsZero() {
		t = t + " (" + note.Due.Format("Monday Jan _2 3:04PM") + ")"
	}

	if lastNote == nil || lastNote.CapturedAt.Day() != note.CapturedAt.Day() {
		t = t + " " + "<strong>" + note.CapturedAt.Format("Jan _2") + "</strong>"
	}

	return t
}

func postNote(c *httpx.Ctx) error {
	r := c.Request()

	filename := r.Header.Get("Filename")
	name, id, err := parseNoteFilename(filename)
	if err != nil {
		return fmt.Errorf("invalid 'Filename' header")
	}

	slog.Info("new voice note", "name", name, "size", r.ContentLength)

	if r.ContentLength < 14_000 {
		// 14kb is about 1 second of working audio
		return nil
	}

	f, err := os.Create(filepath.Join(voiceNotesDir, name))
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = io.Copy(f, r.Body); err != nil {
		return err
	}

	go func() {
		mu.Lock()
		defer mu.Unlock()

		if _, exists := notes[id]; exists {
			return
		}

		note := &Note{
			ID:         id,
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
			CapturedAt: time.UnixMilli(int64(id)),
			Filename:   name,
		}

		notes[id] = note
		if err := process(context.Background(), note); err != nil {
			slog.Error("transcribing voice note", "error", err)
		}
	}()

	return nil
}

func deleteNote(c *httpx.Ctx) error {
	r := c.Request()
	rawID := r.PathValue("id")
	id, _ := strconv.Atoi(rawID)

	mu.Lock()
	defer mu.Unlock()

	note, exists := notes[id]
	if !exists {
		return nil
	}

	err := os.Remove(filepath.Join(voiceNotesDir, note.Filename))
	if err != nil {
		return nil
	}
	delete(notes, id)

	return nil
}

func editNote(c *httpx.Ctx) error {
	r := c.Request()
	rawID := r.PathValue("id")
	id, _ := strconv.Atoi(rawID)
	text := r.FormValue("text")
	if text == "" {
		return nil
	}

	mu.Lock()
	defer mu.Unlock()

	note, exists := notes[id]
	if !exists {
		return nil
	}

	note.Text = text

	return nil
}

func processNote(c *httpx.Ctx) error {
	mu.Lock()
	mu.Unlock()

	rawID := c.Request().PathValue("id")
	id, _ := strconv.Atoi(rawID)

	note, ok := notes[id]
	if !ok {
		return nil
	}

	return process(c.Request().Context(), note)
}

func parseNoteFilename(filename string) (name string, id int, err error) {
	name = filepath.Base(filename) // Tasker sends the entire file path
	rawID, _, _ := strings.Cut(name, ".")
	id, err = strconv.Atoi(rawID)
	return
}

func sortMap[K cmp.Ordered, V any](m map[K]V, cmp func(a V, b V) int) []V {
	values := make([]V, len(m))
	i := 0
	for _, v := range m {
		values[i] = v
		i++
	}
	slices.SortFunc(values, cmp)
	return values
}
