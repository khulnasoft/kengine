package reverseproxy

import (
	"github.com/khulnasoft/kengine"
)

// Register kengine module.
func init() {
	kengine.RegisterModule(kengine.Module{
		Name: "http.responders.reverse_proxy",
		New:  func() interface{} { return new(LoadBalanced) },
	})
}
