package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	htmgo "github.com/robertkozin/x/htmgo"
	"github.com/robertkozin/x/internal/httpx"
	"github.com/samber/oops"
)

func pageIndex(c *httpx.Ctx) (err error) {
	ctx := c.Request().Context()
	rows, _ := db.Query(ctx, `select id, filename, recorded_at, text, summary, category, due from notes order by recorded_at desc`)
	notes, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[Note])
	if err != nil {
		return oops.Wrap(err)
	}

	return htmgo.Render(context.Background(), c.Response(), Index{notes: notes})
}

func (i Index) needsDivider(note *Note, lastNote *Note) (string, bool) {
	if lastNote == nil || lastNote.RecordedAt.Day() != note.RecordedAt.Day() {
		return note.RecordedAt.Format("Jan _2"), true
	}
	return "", false
}

func (i Index) NoteTitle(note *Note, lastNote *Note) string {
	t := note.Summary
	if t == "" {
		t = note.Text
	}
	if t == "" {
		t = `(none)`
	} else if len(t) > 150 {
		t = t[:150]
		t = t + "..."
	}

	if note.Category != "" {
		t = fmt.Sprintf("[%s] %s", note.Category, t)
	}

	if due, ok := note.Due.Get(); ok {
		var d string
		if due.Hour() == 0 && due.Minute() == 0 {
			d = due.Format("Mon Jan _2")
		} else if due.Minute() == 0 {
			d = due.Format("Mon Jan _2 3PM")
		} else {
			d = due.Format("Mon Jan _2 3:04PM")
		}
		t = t + " <ins>" + d + "</ins>"
	}

	return t
}
