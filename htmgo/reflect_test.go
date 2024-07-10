package htmgo

import (
	"fmt"
	"io"
	"reflect"
	"testing"
)

type Robert struct {
	Name    string
	Age     int
	Gender  int
	Address Addr
}

type Addr struct {
	Street   string
	Num      string
	Whatever string
}

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()

	b.Run("new generic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r := NewGen[Robert]()
			fmt.Fprint(io.Discard, r)
		}
	})

	rt := reflect.TypeOf(Robert{})

	b.Run("new reflect", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r := reflect.New(rt).Interface().(*Robert)

			fmt.Fprint(io.Discard, r)
		}
	})
}

func yo[A any, B any]() bool {
	return reflect.TypeFor[A]() == reflect.TypeFor[B]()
}

func TestEq(t *testing.T) {
	fmt.Println(yo[int, int]())
	fmt.Println(yo[int, string]())
	fmt.Println(yo[string, []byte]())
	fmt.Println(yo[Robert, Robert]())
}

func NewGen[T any]() *T {
	return new(T)
}
