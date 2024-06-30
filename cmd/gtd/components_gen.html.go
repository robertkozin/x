package main

import (
	"context"
	"io"

	"github.com/robertkozin/x/htmgo"
)

type Index struct{}

func (props Index) RenderWriter(ctx context.Context, w io.Writer) error {
	return htmgo.RenderWriter(ctx, w, props.Render)
}

func (props Index) Render(ctx context.Context, w *htmgo.Writer) error {
	w.Html("\n<html lang=\"en\">\n<head>\n    <meta charset=\"utf-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n    <title>GTD</title>\n</head>\n<body>\n<h1>Hello, world! WOW</h1>\n</body>\n</html>\n")
	return w.Err()
}
