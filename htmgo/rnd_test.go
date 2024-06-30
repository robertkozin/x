package htmgo

import (
	"fmt"
	"strings"
	"testing"
)

var x = `<!DOCTYPE html>
<html lang="en">
<head>
</head>
<body>
	<h1>Hey</h1>
	<ol>
		<!-- ayyy lmao -->
		<img src=""/>
		<li>Hey
		<li>Hey
	</ol>
</body>
</html>`

var y = `<div>Hey</div>`

func TestRnd(t *testing.T) {
	z := NewTokenizer(strings.NewReader(x))
	for {
		tt := z.Next()
		data := z.buf[z.data.start:z.data.end]
		fmt.Printf("%s: %q, %q\n", tt.String(), z.Raw(), data)
		switch tt {
		case ErrorToken:
			return
		case TextToken:
		case StartTagToken:
			tag := string(data)
			if tag == "ol" {
				fmt.Println("START OL")
				do(z, tag)
				fmt.Println("END OL")
			}
		case EndTagToken:
		case SelfClosingTagToken:
		case CommentToken:
		case DoctypeToken:
		}
	}
}

func do(z *Tokenizer, end string) {
	for {
		tt := z.Next()
		data := z.buf[z.data.start:z.data.end]
		fmt.Printf("%s: %q, %q\n", tt.String(), z.Raw(), data)
		switch tt {
		case ErrorToken:
			return
		case TextToken:
		case StartTagToken:
			tag := string(data)
			if tag == "ol" {
				fmt.Println("START OL")
				do(z, tag)
				fmt.Println("END OL")
			}
		case EndTagToken:
			tag := string(data)
			if end == tag {
				return
			}
		case SelfClosingTagToken:
		case CommentToken:
		case DoctypeToken:
		}
	}
}
