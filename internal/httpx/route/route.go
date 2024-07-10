package route

import (
	"net/http"
	"reflect"
	"unsafe"
)

type HttpWrapper interface {
	WrapHttp(w http.ResponseWriter, r *http.Request)
	Request() *http.Request
	Response() http.ResponseWriter
}

type Router[T any] struct {
	Mux    *http.ServeMux
	prefix string
	M      []func(next func(next *T) error) func(c *T) error // Middleware
}

func (r *Router[T]) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Mux.ServeHTTP(w, req)
}

func New[Base any, BaseWrapper interface {
	*Base
	HttpWrapper
}]() *Router[Base] {
	return &Router[Base]{Mux: new(http.ServeMux)}
}

func Use[Base any, BaseWrapper interface {
	*Base
	HttpWrapper
}](r *Router[Base], middleware func(next func(*Base) error) func(*Base) error) {
	r.M = append(r.M, middleware)
}

func Handle[Base any, Extend any, ExtendWrapper interface {
	*Extend
	HttpWrapper
}](r *Router[Base], method, pattern string, handler func(ctx *Extend) error) {
	mustValidateWrap[Base, Extend]()

	middleware := r.M
	chain := func(ctx *Base) (err error) {
		extend := (*Extend)(unsafe.Pointer(ctx))
		return handler(extend)
	}
	for i := len(middleware) - 1; i >= 0; i-- {
		chain = middleware[i](chain)
	}

	if method == "" {
		pattern = r.prefix + pattern
	} else {
		pattern = method + " " + r.prefix + pattern
	}

	r.Mux.Handle(pattern, http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		extend := new(Extend)
		ExtendWrapper(extend).WrapHttp(res, req)
		base := (*Base)(unsafe.Pointer(extend))
		_ = chain(base)
	}))
}

func HandleHttp[Base any, BaseWrapper interface {
	*Base
	HttpWrapper
}](r *Router[Base], method, pattern string, handler http.Handler) {
	middleware := r.M
	chain := func(ctx *Base) (err error) {
		w := BaseWrapper(ctx)
		handler.ServeHTTP(w.Response(), w.Request())
		return nil
	}
	for i := len(middleware) - 1; i >= 0; i-- {
		chain = middleware[i](chain)
	}

	if method == "" {
		pattern = r.prefix + pattern
	} else {
		pattern = method + " " + r.prefix + pattern
	}

	r.Mux.Handle(pattern, http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		base := new(Base)
		BaseWrapper(base).WrapHttp(res, req)
		_ = chain(base)
	}))
}
func Group[Base any, BaseWrapper interface {
	*Base
	HttpWrapper
}](r *Router[Base], prefix string, group func(r Router[Base])) {
	group(Router[Base]{
		Mux:    r.Mux,
		prefix: r.prefix + prefix,
		M:      append(r.M),
	})
}

func GroupWrap[Base any, Extend any, ExtendWrapper interface {
	*Extend
	HttpWrapper
}](r *Router[Base], prefix string, group func(r Router[Extend])) {
	mw := func(next func(next *Extend) error) func(c *Extend) error {
		chain := func(ctx *Base) error {
			extend := (*Extend)(unsafe.Pointer(ctx))
			return next(extend)
		}
		for i := len(r.M) - 1; i >= 0; i-- {
			chain = r.M[i](chain)
		}
		return func(ctx *Extend) error {
			base := (*Base)(unsafe.Pointer(ctx))
			return chain(base)
		}
	}

	group(Router[Extend]{
		Mux:    r.Mux,
		prefix: r.prefix + prefix,
		M:      []func(next func(next *Extend) error) func(c *Extend) error{mw},
	})
}

func mustValidateWrap[Base any, Extend any]() {
	base := reflect.TypeFor[Base]()
	extend := reflect.TypeFor[Extend]()

	if base == extend {
		return
	} else if base == extend.Field(0).Type {
		return
	}
	panic("ctx wrap: Base must equal Extend or Extend must embed Base in the first field")
}
