// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package events

import (
	"context"
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/i18n"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

const i18nPrefix = "event"

type definition struct {
	name        string
	description string
}

func (d *definition) With(options ...Option) Builder {
	extended := &builder{
		definition: d,
	}
	extended.options = append(extended.options, options...)
	return extended
}

var defaultOptions = []Option{
	WithVisibility(ttnpb.RIGHT_ALL),
}

func (d *definition) New(ctx context.Context, opts ...Option) Event {
	return d.With(defaultOptions...).New(ctx, opts...)
}

func (d *definition) NewWithIdentifiersAndData(ctx context.Context, ids CombinedIdentifiers, data interface{}) Event {
	return d.With(defaultOptions...).NewWithIdentifiersAndData(ctx, ids, data)
}

func (d *definition) BindData(data interface{}) Builder {
	return d.With(defaultOptions...).BindData(data)
}

// Definitions of registered events.
var definitions = make(map[string]*definition)

// defineSkip registers an event and returns its definition.
// The argument skip is the number of stack frames to ascend, with 0 identifying the caller of defineSkip.
func defineSkip(name, description string, skip uint, opts ...Option) Builder {
	if definitions[name] != nil {
		panic(fmt.Errorf("Event %q already defined", name))
	}
	def := &definition{
		name:        name,
		description: description,
	}
	definitions[name] = def

	i18n.Define(fmt.Sprintf("%s:%s", i18nPrefix, name), description).SetSource(1 + skip)
	initMetrics(name)

	var b Builder = def
	if len(opts) > 0 {
		b = b.With(opts...)
	}
	return b
}

// Define a registered event.
func Define(name, description string, opts ...Option) Builder {
	return defineSkip(name, description, 1, opts...)
}

// DefineFunc generates a function, which returns a Definition with specified name and description.
// Most callers should be using Define - this function is only useful for helper functions.
func DefineFunc(name, description string, opts ...Option) func() Builder {
	return func() Builder {
		return defineSkip(name, description, 1, opts...)
	}
}
