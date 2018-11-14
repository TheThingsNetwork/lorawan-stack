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

package applicationserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	errDuplicateIdentifiers = errors.DefineAlreadyExists("duplicate_identifiers", "identifiers already exists")
)

// DeviceRegistry is a store for end devices.
type DeviceRegistry interface {
	// Get returns the end device by its identifiers.
	Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error)
	// Set creates, updates or deletes the end device by its identifiers.
	Set(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
}

// LinkRegistry is a store for application links.
type LinkRegistry interface {
	// Get returns the link by the application identifiers.
	Get(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttnpb.ApplicationLink, error)
	// Range ranges the links and calls the callback function, until false is returned.
	Range(ctx context.Context, paths []string, f func(context.Context, ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationLink) bool) error
	// Set creates, updates or deletes the link by the application identifiers.
	Set(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string, f func(*ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error)) (*ttnpb.ApplicationLink, error)
}
