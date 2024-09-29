package kenginescript

import (
	"net/http"

	kenginescript "github.com/khulnasoft/kengine/pkg/kenginescript/lib"
	"go.starlark.net/starlark"
)

// MatcherEnv sets up the global context for the matcher kenginescript environment.
func MatcherEnv(r *http.Request) starlark.StringDict {
	env := make(starlark.StringDict)
	env["req"] = kenginescript.HTTPRequest{Req: r}
	env["time"] = kenginescript.Time{}
	env["regexp"] = kenginescript.Regexp{}

	return env
}
