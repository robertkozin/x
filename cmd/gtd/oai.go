package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/samber/oops"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"log/slog"
	"path/filepath"
	"time"
)

func process(ctx context.Context, note *Note) error {
	if err := transcribe(ctx, note); err != nil {
		oops.Wrap(err)
	}
	return oops.Wrap(extract(ctx, note))
}

func transcribe(ctx context.Context, note *Note) error {
	res, err := oai.CreateTranscription(ctx, openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: filepath.Join(voiceNotesDir, note.Filename),
		Language: "en",
	})
	if err != nil {
		return oops.With("file", note.Filename).Wrapf(err, "create transcription")
	}

	note.Text = res.Text

	slog.Info("transcribed audio", "file", note.Filename, "text", note.Text)

	return nil
}

type AddNoteDataArgs struct {
	Summary string `json:"summary"`
	DueDate string `json:"due_date"`
	DueTime string `json:"due_time"`
	Topic   string `json:"topic"`
}

func extract(ctx context.Context, note *Note) error {
	params := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"summary": {
				Type:        jsonschema.String,
				Description: "Actionable summary of the note",
			},
			"due_date": {
				Type:        jsonschema.String,
				Description: "The optional due date of the note. Formatted as YYYY-MM-DD",
			},
			"due_time": {
				Type:        jsonschema.String,
				Description: "The optional due time of the note. Formatted like 12:43PM",
			},
			"topic": {
				Type:        jsonschema.String,
				Description: "A one word subject or topic of the text",
			},
		},
		Required: []string{"topic"},
	}

	// TODO: Existing topics

	tool := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "add_note_data",
			Description: "Attach additional data to a transcription of an audio note",
			Parameters:  params,
		},
	}

	prompt := fmt.Sprintf(`Please extract the following information from the transcribed voice note:

1. Summary: A short, actionable title summarizing the task if needed.
2. Topic/Subject: A one word main topic or subject that the note is about. Choose from the following if applicable: Health, Chores, Tasks, Ideas, Notes, Reminders, Lola, Friends, Work, Projects.
3. Due Date and Time (if mentioned): Any specified due date and time for the task. The current date and time is %s %s.
`,
		time.Now().Format(time.DateOnly),
		time.Now().Format(time.Kitchen),
	)

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: note.Text,
			},
		},
		Tools: []openai.Tool{tool},
		ToolChoice: openai.ToolChoice{
			Type: openai.ToolTypeFunction,
			Function: openai.ToolFunction{
				Name: "add_note_data",
			},
		},
	}

	resp, err := oai.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return oops.Wrap(err)
	} else if len(resp.Choices) == 0 || len(resp.Choices[0].Message.ToolCalls) == 0 {
		return oops.With("choices", resp.Choices).Errorf("missing choices or tool calls")
	}

	call := resp.Choices[0].Message.ToolCalls[0].Function

	var args AddNoteDataArgs
	err = json.Unmarshal([]byte(call.Arguments), &args)
	if err != nil {
		return oops.With("args", call.Arguments).Wrapf(err, "unmarshal tool call args")
	}

	var due time.Time
	if args.DueDate != "" && args.DueTime == "" {
		due, err = time.Parse(time.DateOnly, args.DueDate)
	} else if args.DueDate != "" && args.DueTime != "" {
		due, err = time.Parse(time.DateOnly+" "+time.Kitchen, args.DueDate+" "+args.DueTime)
	} else if args.DueDate == "" && args.DueTime != "" {
		today := time.Now()
		var dueTime time.Time
		dueTime, err = time.Parse(time.Kitchen, args.DueTime)
		due = time.Date(today.Year(), today.Month(), today.Day(), dueTime.Hour(), dueTime.Minute(), 0, 0, time.UTC)
	}
	if err != nil {
		return oops.With("date", args.DueDate, "time", args.DueTime).Wrap(err)
	}

	note.Summary = args.Summary
	note.Topic = args.Topic
	note.Due = due
	return nil
}
