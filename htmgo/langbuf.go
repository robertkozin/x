package htmgo

import (
	"fmt"
	"io"
)

const (
	langHtml = iota
	langGo
	langNone
)

type LangBuf struct {
	w    io.Writer
	buf  []byte
	lang int
}

func (b *LangBuf) Gof(format string, args ...any) {
	b.flush(langGo)
	b.buf = fmt.Appendf(b.buf, format, args...)
	b.lang = langGo
}

func (b *LangBuf) Htmlf(format string, args ...any) {
	b.flush(langHtml)
	b.buf = fmt.Appendf(b.buf, format, args...)
	b.lang = langHtml
}

func (b *LangBuf) Flush() {
	b.flush(langNone)
}

func (b *LangBuf) flush(lang int) {
	if b.lang == lang || len(b.buf) == 0 {
		return
	}

	switch b.lang {
	case langHtml:
		fmt.Fprintf(b.w, "w.Html(%q)\n", b.buf) // TODO: chunk on 255 length something like that
	case langGo:
		b.w.Write(b.buf)
	case langNone:
		break
	}

	b.buf = b.buf[:0]
}
