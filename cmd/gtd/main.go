package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var started = time.Now()

var index = `
Hello, World!
`

func main() {
	log.Println("Hello, World!")

	Must1(os.MkdirAll("./voice-notes", os.ModePerm))

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "index.html", started, strings.NewReader(index))
	})

	http.HandleFunc("POST /voice-note", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("POST /voice-note size=%d\n", r.ContentLength)
		fname := fmt.Sprintf("%d.mp4", time.Now().Unix())
		f := Must2(os.Create(filepath.Join("./voice-notes", fname)))
		defer f.Close()

		Must2(io.Copy(f, r.Body))

		log.Printf("Saved %s\n", fname)
	})

	log.Println("Listening on port 8080")
	if err := http.ListenAndServe(":8080", nil); !errors.Is(err, http.ErrServerClosed) {
		log.Println(err)
		os.Exit(1)
	}
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
