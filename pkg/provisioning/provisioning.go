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

// Package provisioning provides a registry and implementations of vendor-specific device provisioners.
package provisioning

import (
	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

// Provisioner is a device provisioner based on vendor-specific data.
type Provisioner interface {
	// UniqueID returns the vendor-specific unique ID for the given entry.
	UniqueID(entry *pbtypes.Struct) (string, error)
}

var (
	registry = map[string]Provisioner{}

	errEntry = errors.DefineInvalidArgument("entry", "invalid entry")
)

// Get returns the provisioner by ID.
func Get(id string) Provisioner {
	return registry[id]
}

// Register registers the given provisioner.
// Existing registrations with the same ID will be overwritten.
// This function is not goroutine-safe.
func Register(id string, p Provisioner) {
	registry[id] = p
}
