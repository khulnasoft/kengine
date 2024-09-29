package kenginebrotli

import (
	"fmt"

	"github.com/andybalholm/brotli"
	"github.com/khulnasoft/kengine"
	"github.com/khulnasoft/kengine/modules/kenginehttp/encode"
)

func init() {
	kengine.RegisterModule(kengine.Module{
		Name: "http.encoders.brotli",
		New:  func() interface{} { return new(Brotli) },
	})
}

// Brotli can create brotli encoders. Note that brotli
// is not known for great encoding performance.
type Brotli struct {
	Quality *int `json:"quality,omitempty"`
}

// Validate validates b's configuration.
func (b Brotli) Validate() error {
	if b.Quality != nil {
		quality := *b.Quality
		if quality < brotli.BestSpeed {
			return fmt.Errorf("quality too low; must be >= %d", brotli.BestSpeed)
		}
		if quality > brotli.BestCompression {
			return fmt.Errorf("quality too high; must be <= %d", brotli.BestCompression)
		}
	}
	return nil
}

// NewEncoder returns a new brotli writer.
func (b Brotli) NewEncoder() encode.Encoder {
	quality := brotli.DefaultCompression
	if b.Quality != nil {
		quality = *b.Quality
	}
	return brotli.NewWriterLevel(nil, quality)
}

// Interface guards
var (
	_ encode.Encoding  = (*Brotli)(nil)
	_ kengine.Validator = (*Brotli)(nil)
)