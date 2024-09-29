package main

import (
	kenginecmd "github.com/khulnasoft/kengine/cmd"

	// this is where modules get plugged in
	_ "github.com/khulnasoft/kengine/modules/kenginehttp"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/kenginelog"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/encode"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/encode/brotli"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/encode/gzip"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/encode/zstd"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/fileserver"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/headers"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/markdown"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/requestbody"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/reverseproxy"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/rewrite"
	_ "github.com/khulnasoft/kengine/modules/kenginetls"
	_ "github.com/khulnasoft/kengine/modules/kenginetls/standardstek"
)

func main() {
	kenginecmd.Main()
}
