package main

import (
	"context"

	"github.com/robertkozin/x/htmgo"
)

type HelloWorld struct{ people []Person }

func (props HelloWorld) Render(ctx context.Context, w *htmgo.Writer) error {
	w.Html("<h1>Hello, ")
	w.Print(props.people)
	w.Html("</h1>")
	return w.Err()
}
