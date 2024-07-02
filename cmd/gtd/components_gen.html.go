package main

import (
	"context"
	"io"

	"github.com/robertkozin/x/htmgo"
)

type Index struct {
	notes    []*VoiceNote
	lastNote *VoiceNote
}

func (props Index) RenderWriter(ctx context.Context, w io.Writer) error {
	return htmgo.RenderWriter(ctx, w, props.Render)
}

func (props Index) Render(ctx context.Context, w *htmgo.Writer) error {
	w.Html("\n<html lang=\"en\">\n<head>\n    <meta charset=\"utf-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n    <title>GTD</title>\n    <link rel=\"stylesheet\" href=\"https://cdnjs.cloudflare.com/ajax/libs/concrete.css/3.0.0/concrete.min.css\">\n    <style>\n        * {\n            min-width: 0;\n        }\n\n        main {\n            max-width: 600px !important;\n        }\n\n        details {\n            padding: 13px 0;\n        }\n\n        details[open] summary {\n            opacity: 0.6;\n        }\n\n        details + details {\n            border-top: 1px solid var(--fg);\n        }\n\n        details summary {\n            cursor: pointer;\n            /*font-size: medium;*/\n            user-select: none;\n        }\n\n        audio {\n            width: 100%;\n        }\n    </style>\n</head>\n\n<body>\n<main>\n    <h1>Notes</h1>\n    <hr>\n\n<!--    <label>-->\n<!--        Volume-->\n<!--        <input type=\"range\" value=\"0.3\" max=\"1\" min=\"0\" step=\"0.01\" onchange=\"vol(this.value)\">-->\n<!--    </label>-->\n\n\n    <div style=\"display: flex; flex-direction: column; gap: 5px;\">")
	for _, note := range props.notes {
		w.Html("\n        <details")
		w.Attr("data-id", note.ID)
		w.Html(">\n            <summary>")
		w.PrintString(props.NoteTitle(note, props.lastNote))
		w.Html("\n            </summary>\n            <p>")
		w.Print(note.CapturedAt.Format("Monday Jan _2 3:04PM"))
		w.Html("</p>")
		if note.Text != "" {
			w.Html("\n            <p data-trans=\"\">")
			w.PrintString(note.Text)
			w.Html("\n            </p>")
		} else {
			w.Html("\n            <p data-trans=\"\">(none)</p>")
		}
		w.Html("\n\n\n            <p>\n                <button data-edit=\"\">Edit</button>\n                <button data-delete=\"\">Delete</button>\n            </p>\n            <p>\n                <audio controls=\"\" preload=\"none\"")
		w.Attr("src", "/voice-note/", note.Filename)
		w.Html(">\n                </audio>\n            </p>\n        </details>")
		props.lastNote = note
		w.Html("\n        <template>\n        </template>")
	}
	w.Html("\n    </div>\n</main>\n<script>\n\n    function vol(level) {\n        document.querySelectorAll(\"audio\").forEach(e => e.volume = level)\n    }\n\n    vol(0.30)\n\n    document.querySelectorAll(\"[data-edit]\").forEach(el => {\n        el.addEventListener(\"click\", e => {\n            let det = e.currentTarget.closest(\"details\")\n            let trans = det.querySelector(\"[data-trans]\").innerText\n            console.log(\"new\", prompt(\"Edit\", trans)) // will return null if cancel\n        })\n    })\n\n    document.querySelectorAll(\"[data-delete]\").forEach(el => {\n        el.addEventListener(\"click\", e => {\n            confirm(\"Delete?\")\n        })\n    })\n</script>\n</body>\n</html>\n")
	return w.Err()
}
