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

package ttnpb

import (
	"context"
	"fmt"
)

// UniqueIdentifier provides functionality to get the unique identifier.
type UniqueIdentifier interface {
	UniqueID(context.Context) string
}

// UniqueID returns the unique identifier.
func (ids UserIdentifiers) UniqueID(_ context.Context) string { return ids.UserID }

// UniqueID returns the unique identifier.
func (ids ApplicationIdentifiers) UniqueID(_ context.Context) string { return ids.ApplicationID }

// UniqueID returns the unique identifier.
func (ids GatewayIdentifiers) UniqueID(_ context.Context) string { return ids.GatewayID }

// UniqueID returns the unique identifier.
func (ids EndDeviceIdentifiers) UniqueID(ctx context.Context) string {
	return fmt.Sprintf("%v:%v", ids.ApplicationIdentifiers.UniqueID(ctx), ids.DeviceID)
}

// UniqueID returns the unique identifier.
func (ids ClientIdentifiers) UniqueID(_ context.Context) string { return ids.ClientID }

// UniqueID returns the unique identifier.
func (ids OrganizationIdentifiers) UniqueID(_ context.Context) string { return ids.OrganizationID }

// GatewayIdentifiersFromUniqueID returns the gateway identifiers that could be extracted from the unique ID.
func GatewayIdentifiersFromUniqueID(uniqueID string) (*GatewayIdentifiers, error) {
	return &GatewayIdentifiers{
		GatewayID: uniqueID,
	}, nil
}
