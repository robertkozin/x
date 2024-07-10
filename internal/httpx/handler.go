package httpx

import (
	"net/http"
)

type Ctx struct {
	response http.ResponseWriter
	request  *http.Request
}

func (c *Ctx) WrapHttp(w http.ResponseWriter, r *http.Request) {
	c.response = w
	c.request = r
}

func (c *Ctx) Request() *http.Request {
	return c.request
}

func (c *Ctx) Response() http.ResponseWriter {
	return c.response
}
