package htmgo

import (
	"bytes"
	"unicode"
)

type Attr struct {
	key []byte
	val []byte
}

type Trans struct {
	z *Tokenizer
	w LangBuf

	imports []string

	tag           string
	indent        []byte
	data          []byte
	directives    [20][]byte
	hasDirectives bool
	attrs         []Attr
}

func (p *Trans) Next() TokenType {
	tt := p.z.Next()

	p.data = p.z.buf[p.z.data.start:p.z.data.end]

	p.tag = ""
	//p.indent = nil
	for i, _ := range p.directives {
		p.directives[i] = nil
	}
	p.hasDirectives = false
	p.attrs = p.attrs[:0]

	if tt == TextToken {
		padIdx := bytes.LastIndexFunc(p.data, func(r rune) bool {
			return !unicode.IsSpace(r)
		})
		if len(p.data) == 0 {

		} else if padIdx > -1 {
			p.indent = p.data[padIdx+1:]
			p.data = p.data[:padIdx+1]
		} else if padIdx == -1 && len(p.data) > 0 {
			p.indent = p.data
			p.data = p.data[:0]
		}
	}

	if tt == StartTagToken || tt == SelfClosingTagToken || tt == EndTagToken {
		var (
			buf      = p.z.buf
			attrs    = p.z.attr
			attr     [2]span
			key, val []byte
		)

		p.tag = string(p.data)

		for i := 0; i < len(attrs); i++ {
			attr = attrs[i]
			key = buf[attr[0].start:attr[0].end]
			val = buf[attr[1].start:attr[1].end]

			if d := mapDirectives(key); d != goNone {
				p.directives[d] = val
				p.hasDirectives = true
				continue
			}

			p.attrs = append(p.attrs, Attr{key, val})
		}
	}

	return tt
}

const (
	goNone = iota
	goIf
	goElseIf
	goElse
	goFor
	goPrint
	goVar
	goOmit
	goIgnore
	goComment
	goSlot
	goComponent
	goFields
	goImport
	goPrintString
	goExpr
	goTimeFormat
)

func mapDirectives(b []byte) int {
	if !bytes.HasPrefix(b, []byte("go-")) {
		return goNone
	}
	b = bytes.TrimPrefix(b, []byte("data-"))
	b = bytes.TrimPrefix(b, []byte("go-"))
	b = bytes.TrimPrefix(b, []byte("go"))
	switch string(b) {
	case "if":
		return goIf
	case "else-if", "elseif":
		return goElseIf
	case "else":
		return goElse
	case "range", "for":
		return goFor
	case "print":
		return goPrint
	case "print-string":
		return goPrintString
	case "var":
		return goVar
	case "omit":
		return goOmit
	case "ignore":
		return goIgnore
	case "comment":
		return goComment
	case "slot":
		return goSlot
	case "component":
		return goComponent
	case "fields":
		return goFields
	case "import":
		return goImport
	case "expr":
		return goExpr
	case "time-format", "timeformat":
		return goTimeFormat
	default:
		return goNone
	}
}

func (p *Trans) DoComponents2() {
	defer p.w.Flush()
	for {
		switch p.Next() {
		case ErrorToken:
			return
		case TextToken, EndTagToken, CommentToken, DoctypeToken, SelfClosingTagToken:
			continue
		case StartTagToken:
			if voidElements[p.tag] || p.directives[goComponent] == nil {
				continue
			}
			p.doOpenTag(p.tag)
		}
	}
}

func (p *Trans) DoComponents(pkg string) {
	// Parse only top level elements with a go-component directive
	p.w.Gof("package %s\n\n", pkg) // TODO: lol
	p.w.Gof(`import (
	"io"
	"fmt"
	"context"
	"github.com/robertkozin/x/htmgo"
	)
	
	`)
	defer p.w.Flush()
	for {
		switch p.Next() {
		case ErrorToken:
			return
		case TextToken, EndTagToken, CommentToken, DoctypeToken, SelfClosingTagToken:
			continue
		case StartTagToken:
			if voidElements[p.tag] || p.directives[goComponent] == nil {
				continue
			}
			p.doOpenTag(p.tag)
		}
	}
}

func (p *Trans) DoRegular(tag string) {
	for {
		switch p.Next() {
		case ErrorToken:
			return
		case TextToken:
			p.doText()
		case StartTagToken, SelfClosingTagToken:
			if p.hasDirectives {
				p.doTag(p.tag)
			} else {
				p.writeTag()
			}
		case EndTagToken:
			if p.tag == tag { // The handler will handle writing the close tag
				return
			}
			p.writeCloseTag()
		case CommentToken:
			p.w.Htmlf("%s<!--%s-->", p.indent, p.data)
		case DoctypeToken:
			p.w.Htmlf("<!DOCTYPE %s>", p.data)
		}
	}
}

func (p *Trans) doText() {
	raw := p.data

	i, i2, ok := hasOpenClose(raw)
	for ok {
		p.w.Htmlf("%s", raw[:i])
		p.w.Gof("w.Print(%s)\n", raw[i+2:i2])
		raw = raw[i2+2:]
		i, i2, ok = hasOpenClose(raw)
	}

	if len(raw) > 0 {
		p.w.Htmlf("%s", raw)
	}
}

func (p *Trans) doTag(tag string) {
	if p.z.tt == SelfClosingTagToken || voidElements[tag] {
		p.doClosedTag(tag)
	} else {
		p.doOpenTag(tag)
	}
}

func (p *Trans) doClosedTag(tag string) {
	p.writeTag()
}

func (p *Trans) doOpenTag(tag string) {
	if p.directives[goImport] != nil {
		p.imports = append(p.imports, string(p.directives[goImport]))
	}

	if p.directives[goComponent] != nil {
		if p.directives[goFields] != nil {
			p.w.Gof("type %s struct {%s}\n\n", p.directives[goComponent], p.directives[goFields])
		} else {
			//p.w.Gof("type %s struct {}\n\n", p.directives[goComponent])
		}

		p.w.Gof(`func (props %[1]s) Render(ctx context.Context, w *htmgo.Writer) error {
	`, p.directives[goComponent])
		defer p.w.Gof("return w.Err()\n}\n\n")
		defer func() { p.w.Htmlf("%s", p.indent) }()
	}

	if p.directives[goVar] != nil {
		p.w.Gof("var(%s)\n", p.directives[goVar])
	}

	if p.directives[goExpr] != nil {
		p.w.Gof("%s\n", p.directives[goExpr])
	}

	if p.directives[goTimeFormat] != nil {
		p.w.Gof("oldFormat := w.TimeFormat(%s)\n", p.directives[goTimeFormat])
		defer p.w.Gof("_ = w.TimeFormat(oldFormat)\n")
	}

	if p.directives[goComment] != nil {
		p.w.Htmlf("<!--")
		defer p.w.Htmlf("-->")
	}

	if p.directives[goIf] != nil { //if
		p.w.Gof("if %s {\n", p.directives[goIf])
		defer p.w.Gof("}\n")
	} else if p.directives[goElseIf] != nil { // elseif
		p.w.buf = bytes.TrimRight(p.w.buf, "\n")
		p.w.Gof("else if %s {\n", p.directives[goElseIf])
		defer p.w.Gof("}\n")
	} else if p.directives[goElse] != nil { // else
		p.w.buf = bytes.TrimRight(p.w.buf, "\n")
		p.w.Gof("else {\n")
		defer p.w.Gof("}\n")
	}

	p.writeTag()
	defer p.writeCloseTag()

	//if p.tag == "template" {
	//	p.w.Htmlf("<!--template-->")
	//	defer p.w.Htmlf("<!--/template-->")
	//} else if p.tag == "slot" {
	//	if len(p.padding) > 0 {
	//		p.w.Htmlf("%s", p.padding)
	//		p.padding = nil
	//	}
	//	p.w.Gof(`htmlgo.CallSlot(ctx, w, "default", func(ctx context.Context, w *htmlgo.Writer) {`)
	//	defer p.w.Gof("})\n")
	//} else if isC {
	//	if len(p.padding) > 0 {
	//		p.w.Htmlf("%s", p.padding)
	//		p.padding = nil
	//	}
	//	p.writeComponent(tag)
	//} else {
	//	p.writeTag(tag)
	//}

	if p.directives[goFor] != nil { //for
		p.w.Gof("for %s {\n", p.directives[goFor])
		defer p.w.Gof("}\n")
	}

	if p.directives[goPrint] != nil { //print
		p.w.Gof("w.Print(%s)\n", p.directives[goPrint])
	}

	if p.directives[goPrintString] != nil {
		p.w.Gof("w.PrintString(%s)\n", p.directives[goPrintString])
	}

	p.DoRegular(tag)
}

func (p *Trans) writeTag() {
	p.w.Htmlf("%s<%s", p.indent, p.tag)
	for _, attr := range p.attrs {
		val := attr.val
		s, e, ok := hasOpenClose(val)
		if ok {
			p.w.Gof("w.Attr(%q", attr.key)
			for ok {
				if len(val[:s]) > 0 {
					p.w.Gof(",%q", val[:s])
				}
				p.w.Gof(",%s", val[s+2:e])
				val = val[e+2:]
				s, e, ok = hasOpenClose(val)
			}
			if len(val) > 0 {
				p.w.Gof(",%s", val)
			}
			p.w.Gof(")\n")

		} else {
			// TODO: escape
			p.w.Htmlf(" %s=\"%s\"", attr.key, attr.val)
		}
	}

	if p.z.tt == SelfClosingTagToken {
		p.w.Htmlf("/")
	}
	p.w.Htmlf(">")
}

func (p *Trans) writeCloseTag() {
	p.w.Htmlf("%s</%s>", p.indent, p.tag)
}

func hasOpenClose(b []byte) (start int, end int, ok bool) {
	start = bytes.Index(b, []byte{'{', '{'})
	end = bytes.Index(b, []byte{'}', '}'}) // TODO: should be b[start:] to prevent thingys that are the same %% %%
	ok = end > start
	return
}
