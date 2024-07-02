package main

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"sync"
	"time"
)

type VoiceNote struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	CapturedAt time.Time
	Order      int    // TODO: make this work, make it initially be captured at
	Filename   string `gorm:"unique"`
	Text       string
}

var transcribeMutex sync.Mutex

func (vn *VoiceNote) Transcribe() {
	if vn.Text != "" {
		return
	}

	transcribeMutex.Lock()
	defer transcribeMutex.Unlock()

	res, err := oai.CreateTranscription(context.Background(), openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: fmt.Sprintf("./voice-note/%s", vn.Filename),
		Language: "en",
	})
	if err != nil {
		log.Println("transcribe error", err)
	}

	db.Model(vn).Update("transcript", res.Text)

	log.Printf("added transcript to %q: %q\n\n", vn.Filename, vn.Text)

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
