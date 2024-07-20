package components

import (
	"context"
	"github.com/robertkozin/x/htmgo"
)

type Button struct{ Name string }

func (props Button) Render(ctx context.Context, w *htmgo.Writer) error {
	w.Html("<button>")
	w.Print(props.Name)
	w.Html("</button>")
	return w.Err()
}
