package main

import (
	"context"

	"github.com/robertkozin/x/htmgo"
)

type Index struct {
	notes    []*Note
	lastNote *Note
}

func (props Index) Render(ctx context.Context, w *htmgo.Writer) error {
	w.Html("\n<html lang=\"en\">\n<head>\n    <meta charset=\"utf-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n    <title>GTD</title>\n    <link rel=\"stylesheet\" href=\"https://cdnjs.cloudflare.com/ajax/libs/concrete.css/3.0.0/concrete.min.css\">\n    <style>\n        * {\n            min-width: 0;\n        }\n\n        main {\n            max-width: 600px !important;\n        }\n\n        details {\n            padding: 13px 0;\n        }\n\n        details[open] summary {\n            opacity: 0.6;\n        }\n\n        details + details {\n            margin-top: 5px;\n        }\n\n        details summary {\n            cursor: pointer;\n            /*font-size: medium;*/\n            user-select: none;\n        }\n\n        .divider {\n            font-size: small;\n            display: flex;\n            align-items: center;\n        }\n\n        .divider::before, .divider::after {\n            flex: 1;\n            content: '';\n            padding: 1px;\n            background-color: var(--fg);\n            margin: 5px;\n        }\n\n        .due {\n            text-decoration: underline;\n        }\n\n        audio {\n            width: 100%;\n        }\n    </style>\n</head>\n\n<body>\n<main>\n    <h1>Notes</h1>\n\n<!--    <label>-->\n<!--        Volume-->\n<!--        <input type=\"range\" value=\"0.3\" max=\"1\" min=\"0\" step=\"0.01\" onchange=\"vol(this.value)\">-->\n<!--    </label>-->")
	oldFormat := w.TimeFormat("Monday Jan _2 3:04PM")
	w.Html("\n\n\n    <div style=\"display: flex; flex-direction: column;\">")
	for _, note := range props.notes {
		if day, needs := props.needsDivider(note, props.lastNote); needs {
			w.Html("\n        <div class=\"divider\">")
			w.Print(day)
			w.Html("</div>")
		}
		props.lastNote = note
		w.Html("\n        <details")
		w.Attr("data-id", note.ID)
		w.Html(">\n            <summary>")
		w.PrintString(props.NoteTitle(note, props.lastNote))
		w.Html("\n            </summary>\n            <p>")
		w.Print(note.RecordedAt)
		w.Html("\n            </p>")
		if note.Text != "" {
			w.Html("\n            <p data-text=\"\">")
			w.PrintString(note.Text)
			w.Html("\n            </p>")
		} else {
			w.Html("\n            <p data-text=\"\">(none)</p>")
		}
		w.Html("\n            <p>\n                <button data-edit=\"\">Edit</button>\n                <button data-delete=\"\">Delete</button>\n                <button data-process=\"\">Process</button>\n            </p>\n            <p>\n                <audio controls=\"\" preload=\"none\">\n                    <source")
		w.Attr("src", "/voice-note/", note.Filename)
		w.Html(" type=\"audio/mp4\">\n                    Your browser does not support the audio element.\n                </audio>\n            </p>\n        </details>")
	}
	w.Html("\n    </div>")
	_ = w.TimeFormat(oldFormat)
	w.Html("\n</main>\n<script>\n\n    function vol(level) {\n        document.querySelectorAll(\"audio\").forEach(e => e.volume = level)\n    }\n\n    vol(0.30)\n\n    document.querySelectorAll(\"[data-edit]\").forEach(el => {\n        el.addEventListener(\"click\", ev => {\n            let note = ev.currentTarget.closest(\"details\")\n            let text = note.querySelector(\"[data-text]\").innerText\n            let newText = prompt(\"Edit\", text)\n            if (newText && text !== newText) {\n                let data = new FormData()\n                data.set(\"text\", newText)\n                fetch(`/note/${note.dataset.id}/edit`, {method: \"POST\", body: data})\n                    .then(() => location.reload())\n            }\n        })\n    })\n\n    document.querySelectorAll(\"[data-process]\").forEach(el => {\n        el.addEventListener(\"click\", ev => {\n            let note = ev.currentTarget.closest(\"details\")\n            fetch(`/note/${note.dataset.id}/process`, {method: \"POST\"}).then(() => location.reload())\n        })\n    })\n\n    document.querySelectorAll(\"[data-delete]\").forEach(el => {\n        el.addEventListener(\"click\", ev => {\n            let note = ev.currentTarget.closest(\"details\")\n            if (confirm(\"Delete?\")) {\n                fetch(`/note/${note.dataset.id}/delete`, {method: \"POST\"})\n                    .then(() => location.reload())\n            }\n        })\n    })\n</script>\n</body>\n</html>\n")
	return w.Err()
}
