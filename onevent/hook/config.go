package hook

import (
	"github.com/khulnasoft/kengine"
)

// Config describes how Hook should be configured and used.
type Config struct {
	ID      string
	Event   kengine.EventName
	Command string
	Args    []string
}

// SupportedEvents is a map of supported events.
var SupportedEvents = map[string]kengine.EventName{
	"startup":   kengine.InstanceStartupEvent,
	"shutdown":  kengine.ShutdownEvent,
	"certrenew": kengine.CertRenewEvent,
}
