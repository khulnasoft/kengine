package markdown

import (
	"net/http"
	"strconv"

	"gopkg.in/russross/blackfriday.v2"

	"github.com/khulnasoft/kengine"
	"github.com/khulnasoft/kengine/modules/kenginehttp"
)

func init() {
	kengine.RegisterModule(kengine.Module{
		Name: "http.middleware.markdown",
		New:  func() interface{} { return new(Markdown) },
	})
}

// Markdown is a middleware for rendering a Markdown response body.
type Markdown struct {
}

func (m Markdown) ServeHTTP(w http.ResponseWriter, r *http.Request, next kenginehttp.Handler) error {
	mrw := &markdownResponseWriter{
		ResponseWriterWrapper: &kenginehttp.ResponseWriterWrapper{ResponseWriter: w},
	}
	return next.ServeHTTP(mrw, r)
}

type markdownResponseWriter struct {
	*kenginehttp.ResponseWriterWrapper
	statusCode  int
	wroteHeader bool
}

func (mrw *markdownResponseWriter) WriteHeader(code int) {
	mrw.statusCode = code
}

func (mrw *markdownResponseWriter) Write(d []byte) (int, error) {
	output := blackfriday.Run(d)
	if !mrw.wroteHeader {
		mrw.Header().Set("Content-Length", strconv.Itoa(len(output)))
		mrw.Header().Set("Content-Type", "text/html; charset=utf-8")
		mrw.WriteHeader(mrw.statusCode)
		mrw.wroteHeader = true
	}
	return mrw.ResponseWriter.Write(output)
}

// Interface guard
var _ kenginehttp.MiddlewareHandler = (*Markdown)(nil)
