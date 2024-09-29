package standard

import (
	// standard Kengine HTTP app modules
	_ "github.com/khulnasoft/kengine/modules/kenginehttp"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/kengineauth"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/encode"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/encode/brotli"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/encode/gzip"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/encode/zstd"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/fileserver"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/headers"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/intercept"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/logging"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/map"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/proxyprotocol"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/push"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/requestbody"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/reverseproxy"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/reverseproxy/fastcgi"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/reverseproxy/forwardauth"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/rewrite"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/templates"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/tracing"
)
