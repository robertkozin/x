package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aarondl/opt/omit"
	"github.com/samber/oops"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"io"
	"time"
)

func transcribe(ctx context.Context, reader io.Reader, filename string) (string, error) {
	res, err := oai.CreateTranscription(ctx, openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: filename,
		Reader:   reader,
		Language: "en",
		// TODO: prompt
	})
	return res.Text, oops.Wrap(err)
}

type extractResponse struct {
	Summary       string              `json:"summary"`
	Category      string              `json:"category"` // required
	DueDate       string              `json:"due_date"`
	DueTime       string              `json:"due_time"`
	ParsedDueDate omit.Val[time.Time] `json:"-"`
}

func extract(ctx context.Context, text string, recordedAt time.Time) (extractResponse, error) {
	if text == "" {
		return extractResponse{}, oops.Errorf("empty extract text")
	}

	prompt := fmt.Sprintf(`Please extract the following information from the transcribed voice note:

1. Summary: A short, actionable title summarizing the task if needed.
2. Category: A one word main topic or subject that the note is about. Choose from the following if applicable: Health, Chores, Tasks, Buy, Ideas, Notes, Reminders, Lola, Friends, Work, Projects.
3. Due date: Any specified due date for the note. The current date is %s.
4. Due time: Any specified due time for the note. The current time is %s.
`, recordedAt.Format(time.DateOnly), recordedAt.Format(time.Kitchen))

	params := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"summary": {
				Type:        jsonschema.String,
				Description: "Actionable summary of the note",
			},
			"category": {
				Type:        jsonschema.String,
				Description: "The category of the note",
			},
			"due_date": {
				Type:        jsonschema.String,
				Description: "The due date specified in the note if there is one. Formatted as YYYY-MM-DD",
			},
			"due_time": {
				Type:        jsonschema.String,
				Description: "The due time specified in the note if there is one. Formatted like 12:43PM",
			},
		},
		Required: []string{"category"},
	}

	tool := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "add_note_data",
			Description: "Attach extracted data from a transcribed audio note",
			Parameters:  params,
		},
	}

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: text,
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

	resp, err := oai.CreateChatCompletion(ctx, req)
	if err != nil {
		return extractResponse{}, oops.Wrap(err)
	} else if len(resp.Choices) == 0 || len(resp.Choices[0].Message.ToolCalls) == 0 {
		return extractResponse{}, oops.With("choices", resp.Choices).Errorf("missing choices or tool calls")
	}

	rawArgs := resp.Choices[0].Message.ToolCalls[0].Function.Arguments

	var args extractResponse
	err = json.Unmarshal([]byte(rawArgs), &args)
	if err != nil {
		return extractResponse{}, oops.With("args", rawArgs).Wrapf(err, "unmarshal tool rawArgs args")
	}

	var due time.Time
	if args.DueDate != "" && args.DueTime == "" {
		due, err = time.Parse(time.DateOnly, args.DueDate)
	} else if args.DueDate != "" && args.DueTime != "" {
		due, err = time.Parse(time.DateOnly+" "+time.Kitchen, args.DueDate+" "+args.DueTime)
	} else if args.DueDate == "" && args.DueTime != "" {
		day := recordedAt
		var dueTime time.Time
		dueTime, err = time.Parse(time.Kitchen, args.DueTime)
		due = time.Date(day.Year(), day.Month(), day.Day(), dueTime.Hour(), dueTime.Minute(), 0, 0, time.UTC)
	}
	err = oops.With("date", args.DueDate, "time", args.DueTime).Wrapf(err, "parse due date/time from tool calll response")
	args.ParsedDueDate = omit.FromCond(due, !due.IsZero())

	return args, err
}
