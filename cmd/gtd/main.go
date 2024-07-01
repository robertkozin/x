package main

import (
	"errors"
	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

var started = time.Now()

var index = `
Hello, World!
`

var (
	db  *gorm.DB
	oai *openai.Client
)

//go:generate go run github.com/robertkozin/x/cmd/htmgo

func main() {
	log.Println("Hello, World!!!")

	Must1(os.MkdirAll("./voice-note", os.ModePerm))

	oai = openai.NewClient(os.Getenv("OPENAI_KEY"))
	db = Must(getDB())

	http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		var notes []VoiceNote
		db.Find(&notes)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		Must1(Index{notes: notes}.RenderWriter(r.Context(), w))
	})

	http.Handle("/voice-note/", http.FileServer(http.Dir(".")))

	http.HandleFunc("POST /new-voice-note", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("POST /voice-note size=%d\n", r.ContentLength)

		filename := r.Header.Get("Filename")
		filename = filepath.Base(filename)
		capturedAt := GetFilenameTimestamp(filename)

		f := Must(os.Create(filepath.Join("./voice-note", filename)))
		defer f.Close()

		Must(io.Copy(f, r.Body))

		vn := VoiceNote{CapturedAt: capturedAt, Filename: filename}

		db.Clauses(
			clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"captured_at", "filename"}),
			},
			clause.Returning{},
		).Select("CapturedAt", "Filename").Create(&vn)

		log.Printf("Saved: %v\n", vn.Filename)

		go func() {
			vn.Transcribe()
		}()
	})

	log.Println("Listening on port 8080")
	if err := http.ListenAndServe("0.0.0.0:8080", nil); !errors.Is(err, http.ErrServerClosed) {
		log.Println(err)
		os.Exit(1)
	}
	log.Println("Done")
}

var matchNumber = regexp.MustCompile(`\d+`)

func GetFilenameTimestamp(filename string) time.Time {
	number := matchNumber.FindString(filename)
	millis := Must(strconv.Atoi(number))
	return time.UnixMilli(int64(millis))
}

func Must1(err error) {
	if err != nil {
		panic("Must1:" + err.Error())
	}
	return
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic("Must:" + err.Error())
	}
	return v
}
