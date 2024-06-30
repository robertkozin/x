package main

import (
	"context"
	"io"

	"github.com/robertkozin/x/htmgo"
)

type Index struct{ notes []VoiceNote }

func (props Index) RenderWriter(ctx context.Context, w io.Writer) error {
	return htmgo.RenderWriter(ctx, w, props.Render)
}

func (props Index) Render(ctx context.Context, w *htmgo.Writer) error {
	w.Html("\n<html lang=\"en\">\n<head>\n    <meta charset=\"utf-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n    <title>GTD</title>\n</head>\n<body>\n<h1>Notes</h1>\n<ol>")
	for _, note := range props.notes {
		w.Html("\n    <li>\n    <audio control=\"\" src=\"\">\n    </audio> <span>")
		w.PrintString(note.Name)
		w.Html(" </span> </li>")
	}
	w.Html("\n</ol>\n</body>\n</html>\n")
	return w.Err()
}
