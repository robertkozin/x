package main

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"time"
)

type VoiceNote struct {
	gorm.Model
	CapturedAt time.Time
	Filename   string `gorm:"unique"`
	Transcript string
}

func (vn *VoiceNote) Transcribe() {
	if vn.Transcript != "" {
		return
	}

	res, err := oai.CreateTranscription(context.Background(), openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: fmt.Sprintf("./voice-note/%s", vn.Filename),
		Language: "en",
	})
	if err != nil {
		log.Println("transcribe error", err)
	}

	db.Model(vn).Update("transcript", res.Text)

	log.Printf("added transcript to %q: %q\n\n", vn.Filename, vn.Transcript)

	return
}

func getDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		return nil, err
	}
	// TODO: sqlite pragma options

	err = db.AutoMigrate(&VoiceNote{})

	return db, err
}
