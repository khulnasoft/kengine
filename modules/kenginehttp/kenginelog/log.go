package kenginelog

import (
	"log"
	"net/http"
	"time"

	"github.com/khulnasoft/kengine"
	"github.com/khulnasoft/kengine/modules/kenginehttp"
)

func init() {
	kengine.RegisterModule(kengine.Module{
		Name: "http.middleware.log",
		New:  func() interface{} { return new(Log) },
	})
}

// Log implements a simple logging middleware.
type Log struct {
	Filename string
	counter  int
}

func (l *Log) ServeHTTP(w http.ResponseWriter, r *http.Request, next kenginehttp.Handler) error {
	start := time.Now()

	// TODO: An example of returning errors
	// return kenginehttp.Error(http.StatusBadRequest, fmt.Errorf("this is a basic error"))
	// return kenginehttp.Error(http.StatusBadGateway, kenginehttp.HandlerError{
	// 	Err:     fmt.Errorf("this is a detailed error"),
	// 	Message: "We had trouble doing the thing.",
	// 	Recommendations: []string{
	// 		"Try reconnecting the gizbop.",
	// 		"Turn off the Internet.",
	// 	},
	// })

	if err := next.ServeHTTP(w, r); err != nil {
		return err
	}

	log.Println("latency:", time.Now().Sub(start), l.counter)

	return nil
}

// Interface guard
var _ kenginehttp.MiddlewareHandler = (*Log)(nil)
