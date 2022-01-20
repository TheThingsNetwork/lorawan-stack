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
	"sort"

	"github.com/gogo/protobuf/proto"
	"go.thethings.network/lorawan-stack/v3/pkg/i18n"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

const i18nPrefix = "event"

type definition struct {
	name        string
	description string
	dataType    proto.Message
}

// Definition describes an event definition.
type Definition interface {
	Name() string
	Description() string
	DataType() proto.Message
}

func (d *definition) Definition() Definition { return d }

func (d definition) Name() string            { return d.name }
func (d definition) Description() string     { return d.description }
func (d definition) DataType() proto.Message { return d.dataType }

func (d *definition) With(options ...Option) Builder {
	extended := &builder{
		definition: d,
	}
	extended.options = append(extended.options, options...)
	return extended
}

var defaultOptions = []Option{
	WithVisibility(ttnpb.Right_RIGHT_ALL),
}

func (d *definition) New(ctx context.Context, opts ...Option) Event {
	return d.With(defaultOptions...).New(ctx, opts...)
}

type EntityIdentifiers interface {
	GetEntityIdentifiers() *ttnpb.EntityIdentifiers
}

func (d *definition) NewWithIdentifiersAndData(ctx context.Context, ids EntityIdentifiers, data interface{}) Event {
	return d.With(defaultOptions...).NewWithIdentifiersAndData(ctx, ids, data)
}

func (d *definition) BindData(data interface{}) Builder {
	return d.With(defaultOptions...).BindData(data)
}

// Definitions of registered events.
var definitions = make(map[string]*definition)

// All returns all defined events, sorted by name.
func All() Builders {
	type definition struct {
		name    string
		builder Builder
	}
	sorted := make([]*definition, 0, len(definitions))
	for name, builder := range definitions {
		sorted = append(sorted, &definition{name: name, builder: builder})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].name < sorted[j].name
	})
	out := make(Builders, len(sorted))
	for i, s := range sorted {
		out[i] = s.builder
	}
	return out
}

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
	for _, opt := range opts {
		if defOpt, ok := opt.(DefinitionOption); ok {
			defOpt.applyToDefinition(def)
		}
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
