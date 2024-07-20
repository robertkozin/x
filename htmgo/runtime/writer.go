package htmgo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Writer struct {
	wr  io.Writer
	buf []byte
	err error

	timeFormat string // TODO: go-timeformat directive!
}

var (
	sp = []byte(" ")
	qt = []byte("'")
)

func (w *Writer) Write(b []byte) (int, error) {
	var n int
	if w.err == nil {
		n, w.err = w.wr.Write(b)
	}
	return n, w.err
}

func (w *Writer) Attr(key string, vals ...any) {
	// Special case for boolean attributes
	if len(vals) == 1 {
		if b, ok := vals[0].(bool); ok && b {
			w.buf = append(w.buf, ' ')
			w.buf = append(w.buf, key...)
			w.flush()
			return
		} else if ok {
			return
		}
	}

	// todo: check if url context etc et
	w.buf = append(w.buf, ' ')
	w.buf = append(w.buf, key...)
	w.buf = append(w.buf, `="`...)
	w.flush()

	for _, val := range vals {
		w.appendPrint(val)
	}
	w.flushSanitize()

	w.buf = append(w.buf, '"')
	w.flush()
}

func (w *Writer) appendPrint(v any) {
	switch x := v.(type) {
	case bool:
		w.buf = strconv.AppendBool(w.buf, x)
	case *string:
		w.buf = append(w.buf, *x...)
	case string:
		w.buf = append(w.buf, x...)
	case int:
		w.buf = strconv.AppendInt(w.buf, int64(x), 10)
	case int64:
		w.buf = strconv.AppendInt(w.buf, int64(x), 10)
	case float64:
		w.buf = strconv.AppendFloat(w.buf, x, 'f', -1, 64)
	case float32:
		w.buf = strconv.AppendFloat(w.buf, float64(x), 'f', -1, 32)
	case time.Time:
		if w.timeFormat == "" {
			w.timeFormat = time.RFC3339
		}
		w.buf = x.AppendFormat(w.buf, w.timeFormat)
	case time.Duration:
		w.buf = append(w.buf, x.String()...)
	case fmt.Stringer:
		w.buf = append(w.buf, x.String()...)
	case fmt.GoStringer:
		w.buf = append(w.buf, x.GoString()...)
	default:
		w.buf = fmt.Append(w.buf, x)
	}

	//w.buf = fmt.Append(w.buf, v)
}

func (w *Writer) TimeFormat(newFormat string) (oldFormat string) {
	oldFormat = w.timeFormat
	w.timeFormat = newFormat
	return
}

func (w *Writer) Reset() {
	w.buf = w.buf[:0]
}

func (w *Writer) flush() {
	//w.flushWrite(w.buf)
	if w.err == nil {
		_, w.err = w.wr.Write(w.buf)
	}
	w.buf = w.buf[:0]
}

var writerPool = sync.Pool{
	New: func() any { return new(Writer) },
}

var noop = func(*Writer) {}
var release = func(w *Writer) { w.buf = w.buf[:0]; w.wr = nil; writerPool.Put(w) }

type Renderer interface {
	Render(ctx context.Context, w *Writer) error
}

func Render(ctx context.Context, w io.Writer, r Renderer) error {
	ww := writerPool.Get().(*Writer)
	ww.wr = w
	defer release(ww)

	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	if _, err := w.Write([]byte("<!doctype html>\n")); err != nil {
		return err
	}

	return r.Render(ctx, ww)
}

func RenderWriter(ctx context.Context, w io.Writer, fn func(context.Context, *Writer) error) error {
	ww := writerPool.Get().(*Writer)
	ww.wr = w
	defer release(ww)
	return fn(ctx, ww)
}

func WrapWriter(_w io.Writer) (*Writer, func(*Writer)) {
	if w, ok := _w.(*Writer); ok {
		return w, noop
	}
	w := writerPool.Get().(*Writer)
	w.wr = _w
	return w, release
}

var encode = [255][]byte{
	'<':  []byte("&lt;"),
	'>':  []byte("&gt;"),
	'"':  []byte("&quot;"),
	'\'': []byte("&#39;"),
	'&':  []byte("&amp;"),
	// TODO: https://wonko.com/post/html-escaping/
}

func (w *Writer) Html(s string) {
	w.buf = append(w.buf, s...)
	w.flush()
}

func (w *Writer) PrintString(s string) {
	w.buf = append(w.buf, s...)
	if w.err == nil {
		_, w.err = w.wr.Write(w.buf)
	}
	w.buf = w.buf[:0]
}

func (w *Writer) Print(v any) {
	w.appendPrint(v)
	if w.err == nil {
		_, w.err = w.wr.Write(w.buf)
	}
	w.buf = w.buf[:0]
}

func (w *Writer) flushWrite(b []byte) {
	if w.err == nil {
		_, w.err = w.wr.Write(b)
	}
}

func (w *Writer) flushSanitize() {
	write := w.flushWrite
	j := 0
	for i, c := range w.buf {
		switch c {
		case '<':
			write(w.buf[j:i])
			write(strLT)
			j = i + 1
		case '>':
			write(w.buf[j:i])
			write(strGT)
			j = i + 1
		case '"':
			write(w.buf[j:i])
			write(strQuot)
			j = i + 1
		case '\'':
			write(w.buf[j:i])
			write(strApos)
			j = i + 1
		case '&':
			write(w.buf[j:i])
			write(strAmp)
			j = i + 1
		}
	}
	write(w.buf[j:])
	w.buf = w.buf[:0]
	return
}

func (w *Writer) Err() error {
	return w.err
}

var (
	strLT   = []byte("&lt;")
	strGT   = []byte("&gt;")
	strQuot = []byte("&quot;")
	strApos = []byte("&#39;")
	strAmp  = []byte("&amp;")
)
