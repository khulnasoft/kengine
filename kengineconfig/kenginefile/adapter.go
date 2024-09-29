// Copyright 2015 Matthew Holt and The Kengine Authors
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

package kenginefile

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/khulnasoft/kengine"
	"github.com/khulnasoft/kengine/kengineconfig"
)

// Adapter adapts Kenginefile to Kengine JSON.
type Adapter struct {
	ServerType ServerType
}

// Adapt converts the Kenginefile config in body to Kengine JSON.
func (a Adapter) Adapt(body []byte, options map[string]any) ([]byte, []kengineconfig.Warning, error) {
	if a.ServerType == nil {
		return nil, nil, fmt.Errorf("no server type")
	}
	if options == nil {
		options = make(map[string]any)
	}

	filename, _ := options["filename"].(string)
	if filename == "" {
		filename = "Kenginefile"
	}

	serverBlocks, err := Parse(filename, body)
	if err != nil {
		return nil, nil, err
	}

	cfg, warnings, err := a.ServerType.Setup(serverBlocks, options)
	if err != nil {
		return nil, warnings, err
	}

	// lint check: see if input was properly formatted; sometimes messy files parse
	// successfully but result in logical errors (the Kenginefile is a bad format, I'm sorry)
	if warning, different := FormattingDifference(filename, body); different {
		warnings = append(warnings, warning)
	}

	result, err := json.Marshal(cfg)

	return result, warnings, err
}

// FormattingDifference returns a warning and true if the formatted version
// is any different from the input; empty warning and false otherwise.
// TODO: also perform this check on imported files
func FormattingDifference(filename string, body []byte) (kengineconfig.Warning, bool) {
	// replace windows-style newlines to normalize comparison
	normalizedBody := bytes.Replace(body, []byte("\r\n"), []byte("\n"), -1)

	formatted := Format(normalizedBody)
	if bytes.Equal(formatted, normalizedBody) {
		return kengineconfig.Warning{}, false
	}

	// find where the difference is
	line := 1
	for i, ch := range normalizedBody {
		if i >= len(formatted) || ch != formatted[i] {
			break
		}
		if ch == '\n' {
			line++
		}
	}
	return kengineconfig.Warning{
		File:    filename,
		Line:    line,
		Message: "Kenginefile input is not formatted; run 'kengine fmt --overwrite' to fix inconsistencies",
	}, true
}

// Unmarshaler is a type that can unmarshal Kenginefile tokens to
// set itself up for a JSON encoding. The goal of an unmarshaler
// is not to set itself up for actual use, but to set itself up for
// being marshaled into JSON. Kenginefile-unmarshaled values will not
// be used directly; they will be encoded as JSON and then used from
// that. Implementations _may_ be able to support multiple segments
// (instances of their directive or batch of tokens); typically this
// means wrapping parsing logic in a loop: `for d.Next() { ... }`.
// More commonly, only a single segment is supported, so a simple
// `d.Next()` at the start should be used to consume the module
// identifier token (directive name, etc).
type Unmarshaler interface {
	UnmarshalKenginefile(d *Dispenser) error
}

// ServerType is a type that can evaluate a Kenginefile and set up a kengine config.
type ServerType interface {
	// Setup takes the server blocks which contain tokens,
	// as well as options (e.g. CLI flags) and creates a
	// Kengine config, along with any warnings or an error.
	Setup([]ServerBlock, map[string]any) (*kengine.Config, []kengineconfig.Warning, error)
}

// UnmarshalModule instantiates a module with the given ID and invokes
// UnmarshalKenginefile on the new value using the immediate next segment
// of d as input. In other words, d's next token should be the first
// token of the module's Kenginefile input.
//
// This function is used when the next segment of Kenginefile tokens
// belongs to another Kengine module. The returned value is often
// type-asserted to the module's associated type for practical use
// when setting up a config.
func UnmarshalModule(d *Dispenser, moduleID string) (Unmarshaler, error) {
	mod, err := kengine.GetModule(moduleID)
	if err != nil {
		return nil, d.Errf("getting module named '%s': %v", moduleID, err)
	}
	inst := mod.New()
	unm, ok := inst.(Unmarshaler)
	if !ok {
		return nil, d.Errf("module %s is not a Kenginefile unmarshaler; is %T", mod.ID, inst)
	}
	err = unm.UnmarshalKenginefile(d.NewFromNextSegment())
	if err != nil {
		return nil, err
	}
	return unm, nil
}

// Interface guard
var _ kengineconfig.Adapter = (*Adapter)(nil)
