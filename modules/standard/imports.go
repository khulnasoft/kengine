package standard

import (
	// standard Kengine modules
	_ "github.com/khulnasoft/kengine/kengineconfig/kenginefile"
	_ "github.com/khulnasoft/kengine/modules/kengineevents"
	_ "github.com/khulnasoft/kengine/modules/kengineevents/eventsconfig"
	_ "github.com/khulnasoft/kengine/modules/kenginefs"
	_ "github.com/khulnasoft/kengine/modules/kenginehttp/standard"
	_ "github.com/khulnasoft/kengine/modules/kenginepki"
	_ "github.com/khulnasoft/kengine/modules/kenginepki/acmeserver"
	_ "github.com/khulnasoft/kengine/modules/kenginetls"
	_ "github.com/khulnasoft/kengine/modules/kenginetls/distributedstek"
	_ "github.com/khulnasoft/kengine/modules/kenginetls/standardstek"
	_ "github.com/khulnasoft/kengine/modules/filestorage"
	_ "github.com/khulnasoft/kengine/modules/logging"
	_ "github.com/khulnasoft/kengine/modules/metrics"
)
