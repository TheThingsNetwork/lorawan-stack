// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

	"go.thethings.network/lorawan-stack/pkg/i18n"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const i18nPrefix = "event"

// Definition of a registered event.
type Definition func(ctx context.Context, identifiers ttnpb.Identifiers, data interface{}) Event

// Definitions of registered events.
// Events that are defined in init() funcs will be collected for translation.
var Definitions = make(map[string]string)

// Define a registered event.
func Define(name, description string) Definition {
	if Definitions[name] != "" {
		panic(fmt.Errorf("Event %s already defined", name))
	}
	i18n.Define(fmt.Sprintf("%s:%s", i18nPrefix, name), description).SetSource(2)
	Definitions[name] = description
	return func(ctx context.Context, identifiers ttnpb.Identifiers, data interface{}) Event {
		return New(ctx, name, identifiers, data)
	}
}
