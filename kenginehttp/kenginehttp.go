// Copyright 2015 KhulnaSoft, Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kenginehttp

import (
	// plug in the server
	_ "github.com/khulnasoft/kengine/kenginehttp/httpserver"

	// plug in the standard directives
	_ "github.com/khulnasoft/kengine/kenginehttp/basicauth"
	_ "github.com/khulnasoft/kengine/kenginehttp/bind"
	_ "github.com/khulnasoft/kengine/kenginehttp/browse"
	_ "github.com/khulnasoft/kengine/kenginehttp/errors"
	_ "github.com/khulnasoft/kengine/kenginehttp/expvar"
	_ "github.com/khulnasoft/kengine/kenginehttp/extensions"
	_ "github.com/khulnasoft/kengine/kenginehttp/fastcgi"
	_ "github.com/khulnasoft/kengine/kenginehttp/gzip"
	_ "github.com/khulnasoft/kengine/kenginehttp/header"
	_ "github.com/khulnasoft/kengine/kenginehttp/index"
	_ "github.com/khulnasoft/kengine/kenginehttp/internalsrv"
	_ "github.com/khulnasoft/kengine/kenginehttp/limits"
	_ "github.com/khulnasoft/kengine/kenginehttp/log"
	_ "github.com/khulnasoft/kengine/kenginehttp/markdown"
	_ "github.com/khulnasoft/kengine/kenginehttp/mime"
	_ "github.com/khulnasoft/kengine/kenginehttp/pprof"
	_ "github.com/khulnasoft/kengine/kenginehttp/proxy"
	_ "github.com/khulnasoft/kengine/kenginehttp/push"
	_ "github.com/khulnasoft/kengine/kenginehttp/redirect"
	_ "github.com/khulnasoft/kengine/kenginehttp/requestid"
	_ "github.com/khulnasoft/kengine/kenginehttp/rewrite"
	_ "github.com/khulnasoft/kengine/kenginehttp/root"
	_ "github.com/khulnasoft/kengine/kenginehttp/status"
	_ "github.com/khulnasoft/kengine/kenginehttp/templates"
	_ "github.com/khulnasoft/kengine/kenginehttp/timeouts"
	_ "github.com/khulnasoft/kengine/kenginehttp/websocket"
	_ "github.com/khulnasoft/kengine/onevent"
)
