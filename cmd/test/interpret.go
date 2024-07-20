package main

import (
	"context"
	"fmt"
	"github.com/robertkozin/x/cmd/components"
	"github.com/robertkozin/x/htmgo"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"os"
	"reflect"
)

var p = `

import (
	"context"

	"github.com/robertkozin/x/htmgo"
	. "_"
)

type HelloWorld struct{ people []Person }

var RenderHelloWorld = func(props HelloWorld, ctx context.Context, w *htmgo.Writer) error {
	return props.Render(ctx, w)
}

func (props HelloWorld) Render(ctx context.Context, w *htmgo.Writer) error {
	w.Html("<h1>Hello, ")
	w.Print(props.people)
	w.Html("</h1>")
	return w.Err()
}

`

type Person string

//go:generate go run github.com/robertkozin/x/cmd/htmgo

//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract .

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func main() {
	//pa := Must(filepath.Abs("."))
	//pa := "/Users/ro/Projects/x/cmd"
	//fs := os.DirFS("/Users/ro/Projects/x/cmd/")
	i := interp.New(interp.Options{GoPath: "/Users/ro/Projects/x"})

	i.Use(stdlib.Symbols)

	Must(0, i.Use(interp.Exports{
		"github.com/robertkozin/x/htmgo/htmgo": map[string]reflect.Value{
			"Writer": reflect.ValueOf(htmgo.Writer{}),
		},
		//"github.com/robertkozin/x/cmd/gtd/gtd": map[string]reflect.Value{
		//	"Button": reflect.ValueOf((*)(nil)),
		//},
		"_/.": map[string]reflect.Value{
			"Person": reflect.ValueOf((*Person)(nil)),
		},
	}))

	props := HelloWorld{people: []Person{"Robert"}}
	ctx := context.Background()
	hw, _ := htmgo.WrapWriter(os.Stdout)

	Must(i.Eval(p))
	//fmt.Println(i.Symbols())

	//v := Must(i.Eval("RenderHelloWorld"))
	v := i.Globals()["RenderHelloWorld"]

	//fn, ok := v.Interface().(func(HelloWorld, context.Context, *htmgo.Writer))

	fmt.Println(reflect.ValueOf((*Person)(nil)).Type().Elem().Name())
	fmt.Println(reflect.ValueOf((*components.Button)(nil)).Type().Elem().PkgPath())

	fmt.Println(v, v.Type())

	v.Call([]reflect.Value{reflect.NewAt(v.Type().In(0), reflect.ValueOf(&props).UnsafePointer()).Elem(), reflect.ValueOf(ctx), reflect.ValueOf(hw)})

	//fmt.Println(reflect.ValueOf(props).Method(0).CanSet())
	//fmt.Println(reflect.ValueOf(props).Method(0).Addr().CanSet())
	//
	//v.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(hw)})

	//fn(props, ctx, hw)

}
