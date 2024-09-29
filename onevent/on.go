package onevent

import (
	"strings"

	"github.com/khulnasoft/kengine"
	"github.com/khulnasoft/kengine/onevent/hook"
	"github.com/google/uuid"
)

func init() {
	// Register Directive.
	kengine.RegisterPlugin("on", kengine.Plugin{Action: setup})
}

func setup(c *kengine.Controller) error {
	config, err := onParse(c)
	if err != nil {
		return err
	}

	// Register Event Hooks.
	err = c.OncePerServerBlock(func() error {
		for _, cfg := range config {
			kengine.RegisterEventHook("on-"+cfg.ID, cfg.Hook)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func onParse(c *kengine.Controller) ([]*hook.Config, error) {
	var config []*hook.Config

	for c.Next() {
		cfg := new(hook.Config)

		if !c.NextArg() {
			return config, c.ArgErr()
		}

		// Configure Event.
		event, ok := hook.SupportedEvents[strings.ToLower(c.Val())]
		if !ok {
			return config, c.Errf("Wrong event name or event not supported: '%s'", c.Val())
		}
		cfg.Event = event

		// Assign an unique ID.
		cfg.ID = uuid.New().String()

		args := c.RemainingArgs()

		// Extract command and arguments.
		command, args, err := kengine.SplitCommandAndArgs(strings.Join(args, " "))
		if err != nil {
			return config, c.Err(err.Error())
		}

		cfg.Command = command
		cfg.Args = args

		config = append(config, cfg)
	}

	return config, nil
}
