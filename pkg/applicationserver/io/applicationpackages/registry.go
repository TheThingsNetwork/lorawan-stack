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

package applicationpackages

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// AssociationRegistry is a registry for application packages data and associations.
type AssociationRegistry interface {
	// Get returns the association by its identifiers.
	Get(ctx context.Context, ids ttnpb.ApplicationPackageAssociationIdentifiers, paths []string) (*ttnpb.ApplicationPackageAssociation, error)
	// List returns all of the associations of the end device.
	List(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string) ([]*ttnpb.ApplicationPackageAssociation, error)
	// Set creates, updates or deletes the association by its identifiers.
	Set(ctx context.Context, ids ttnpb.ApplicationPackageAssociationIdentifiers, gets []string, f func(*ttnpb.ApplicationPackageAssociation) (*ttnpb.ApplicationPackageAssociation, []string, error)) (*ttnpb.ApplicationPackageAssociation, error)
}
