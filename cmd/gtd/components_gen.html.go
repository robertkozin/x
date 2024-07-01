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
	w.Html("\n<html lang=\"en\">\n<head>\n    <meta charset=\"utf-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n    <title>GTD</title>\n    <link rel=\"stylesheet\" href=\"https://cdnjs.cloudflare.com/ajax/libs/concrete.css/3.0.0/concrete.min.css\">\n    <style>\n        main {\n            max-width: 900px !important;\n        }\n    </style>\n</head>\n\n<body>\n<main>\n    <h1>Notes</h1>\n\n    <label>\n        Volume\n        <input type=\"range\" value=\"0.3\" max=\"1\" min=\"0\" step=\"0.01\" onchange=\"vol(this.value)\">\n    </label>\n\n    <table>\n        <thead>\n        <tr>\n            <th>Captured At</th>\n            <th>Voice Note</th>\n            <th>Transcript</th>\n        </tr>\n        </thead>\n\n        <tbody>")
	for _, note := range props.notes {
		w.Html("\n        <tr>\n            <td>")
		w.PrintString(note.CapturedAt.Format("Mon Jan _2 3:04PM"))
		w.Html("\n            </td>\n            <td>\n                <audio controls=\"\" preload=\"none\"")
		w.Attr("src", "/voice-note/", note.Filename)
		w.Html(">\n                </audio>\n            </td>\n            <td>")
		w.PrintString(note.Transcript)
		w.Html("\n            </td>\n        </tr>")
	}
	w.Html("\n        </tbody>\n    </table>\n</main>\n<script>\n    function vol(level) {\n        document.querySelectorAll(\"audio\").forEach(e => e.volume = level)\n    }\n\n    vol(0.30)\n</script>\n</body>\n</html>\n")
	return w.Err()
}
