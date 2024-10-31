package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/robertkozin/x/internal/httpx"
	"github.com/samber/oops"
	"html/template"
	"net/http"
	"time"
)

var (
	tmplComments = template.Must(template.ParseFiles("./cmd/gsites-comments/comments.gohtml"))
)

type CommentsCtx struct {
	httpx.Ctx
	Comments []Comment
	PageId   string
}

type Comment struct {
	ID        string    `db:"id"`
	Content   string    `db:"content"`
	Author    string    `db:"author"`
	CreatedAt time.Time `db:"created_at"`
}

func getComments(c *CommentsCtx) error {
	c.PageId = c.Req().PathValue("pageId")
	ctx := c.Req().Context()

	var err error
	c.Comments, err = dbGetComments(ctx, c.PageId)
	if err != nil {
		return oops.Wrap(err)
	}

	return tmplComments.Execute(c.Response(), c)
}

func postComment(c *httpx.Ctx) error {
	var (
		req     = c.Req()
		ctx     = req.Context()
		pageId  = req.PathValue("pageId")
		name    = req.FormValue("name")
		comment = req.FormValue("comment")
	)

	fmt.Println("post comment", pageId, name, comment)

	err := dbInsertComment(ctx, pageId, name, comment)
	if err != nil {
		return oops.Wrap(err)
	}

	c.Response().Header().Set("Location", fmt.Sprintf("/comments/%s", pageId))
	c.Response().WriteHeader(http.StatusSeeOther)
	return nil
}

func dbGetComments(ctx context.Context, pageId string) ([]Comment, error) {
	q := `select id, content, author, created_at
from comments
where page_id = $1
order by created_at desc`
	rows, _ := db.Query(ctx, q, pageId)
	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[Comment])
}

func dbInsertComment(ctx context.Context, pageId, name, comment string) error {
	q := `insert into comments(page_id, author, content) values ($1, $2, $3)`
	_, err := db.Exec(ctx, q, pageId, name, comment)
	return err
}
