package requestbody

import (
	"net/http"

	"github.com/khulnasoft/kengine"
	"github.com/khulnasoft/kengine/modules/kenginehttp"
)

func init() {
	kengine.RegisterModule(kengine.Module{
		Name: "http.middleware.request_body",
		New:  func() interface{} { return new(RequestBody) },
	})
}

// RequestBody is a middleware for manipulating the request body.
type RequestBody struct {
	MaxSize int64 `json:"max_size,omitempty"`
}

func (rb RequestBody) ServeHTTP(w http.ResponseWriter, r *http.Request, next kenginehttp.Handler) error {
	if r.Body == nil {
		return next.ServeHTTP(w, r)
	}
	if rb.MaxSize > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, rb.MaxSize)
	}
	return next.ServeHTTP(w, r)
}

// Interface guard
var _ kenginehttp.MiddlewareHandler = (*RequestBody)(nil)
