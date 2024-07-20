package eval

import (
	"fmt"
	"github.com/robertkozin/x/htmgo"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"go/ast"
	"go/parser"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func TestEval(t *testing.T) {
	i := interp.New(interp.Options{})

	fs := i.FileSet()

	fp := "/Users/ro/Projects/x/cmd/gtd/"
	entries := Must(os.ReadDir(fp))
	files := []*ast.File{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		} else if !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		f, err := parser.ParseFile(fs, filepath.Join(fp, entry.Name()), nil, 0)
		if err != nil {
			t.Fatal(err, entry.Name())
		}
		files = append(files, f)
	}

	Must(0, i.Use(stdlib.Symbols))

	Must(0, i.Use(interp.Exports{
		"github.com/robertkozin/x/htmgo/htmgo": map[string]reflect.Value{
			"Writer": reflect.ValueOf((*htmgo.Writer)(nil)),
		},
		"main/.": map[string]reflect.Value{
			"Person": reflect.ValueOf((*Person)(nil)),
		},
	}))

	for _, f := range files {
		_, err := i.CompileAST(f)
		if err != nil {
			t.Fatal(err)
		}
	}

	//v := Must(i.Eval(p))

	//fmt.Printf("%#v\n", v)
	//
	//fmt.Println(i.Symbols("main"))
	//
	//sym := i.Symbols("main")
	//
	//hw2 := sym["main"]["HelloWorld"]
	//
	//fmt.Println(hw2, hw2.Type())

	v := Must(i.Eval("HelloWorld.Render"))
	fmt.Println(v, v.Type())

	t.Fail()
}
