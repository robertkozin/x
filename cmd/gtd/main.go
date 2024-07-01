package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var started = time.Now()

var index = `
Hello, World!
`

//go:generate go run ../htmgo/main.go

func main() {
	log.Println("Hello, World!!!")

	Must1(os.MkdirAll("./voice-note", os.ModePerm))

	http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		entries := Must2(os.ReadDir("./voice-note"))
		notes := make([]VoiceNote, len(entries))
		for i, entry := range entries {
			notes[i] = VoiceNote{Name: entry.Name()}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		Must1(Index{notes: notes}.RenderWriter(r.Context(), w))
	})

	http.Handle("/voice-note/", http.FileServer(http.Dir(".")))

	http.HandleFunc("POST /new-voice-note", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("POST /voice-note size=%d\n", r.ContentLength)

		fname := r.Header.Get("Filename")
		fname = filepath.Base(fname)

		f := Must2(os.Create(filepath.Join("./voice-note", fname)))
		defer f.Close()

		Must2(io.Copy(f, r.Body))

		log.Printf("Saved %s\n", fname)
	})

	log.Println("Listening on port 8080")
	if err := http.ListenAndServe("0.0.0.0:8080", nil); !errors.Is(err, http.ErrServerClosed) {
		log.Println(err)
		os.Exit(1)
	}
	log.Println("Done")
}

type VoiceNote struct {
	Name string
}

func Must1(err error) {
	if err != nil {
		panic("Must1:" + err.Error())
	}
	return
}

func Must2[T any](v T, err error) T {
	if err != nil {
		panic("Must2:" + err.Error())
	}
	return v
}
