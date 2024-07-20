package main

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"testing"
	"time"
)

func TestOaiFunc(t *testing.T) {
	oai := openai.NewClient("")

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
`, time.Now().Format(time.DateOnly), time.Now().Format(time.Kitchen))

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "I, uh, should call Chefery.",
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
	if err != nil || len(resp.Choices) != 1 {
		fmt.Printf("Completion error: err:%v len(choices):%v\n", err,
			len(resp.Choices))
		return
	}

	call := resp.Choices[0].Message.ToolCalls[0].Function

	fmt.Printf("func call: %+v\n", call)
}
