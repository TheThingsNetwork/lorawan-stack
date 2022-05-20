// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package gatewayserver

import (
	"context"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GatewayConnectionStatsRegistry stores, updates and cleans up gateway connection stats.
type GatewayConnectionStatsRegistry interface {
	// Get returns connection stats for a gateway.
	Get(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*ttnpb.GatewayConnectionStats, error)
	// Set sets, updates or clears the connection stats for a gateway. Only fields specified in the field mask paths are set.
	Set(ctx context.Context, ids *ttnpb.GatewayIdentifiers, stats *ttnpb.GatewayConnectionStats, paths []string, ttl time.Duration) error
}

// EntityRegistry abstracts the Identity server gateway functions.
type EntityRegistry interface {
	// AssertGatewayRights checks whether the gateway authentication (provied in the context) contains the required rights.
	AssertGatewayRights(ctx context.Context, ids *ttnpb.GatewayIdentifiers, required ...ttnpb.Right) error
	// Get the identifiers of the gateway that has the given EUI registered.
	GetIdentifiersForEUI(ctx context.Context, in *ttnpb.GetGatewayIdentifiersForEUIRequest) (*ttnpb.GatewayIdentifiers, error)
	// Get the gateway with the given identifiers, selecting the fields specified.
	Get(ctx context.Context, in *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error)
	// UpdateAntennas updates the gateway antennas.
	UpdateAntennas(ctx context.Context, ids *ttnpb.GatewayIdentifiers, antennas []*ttnpb.GatewayAntenna) error
	// UpdateAttributes updates the gateway attributes. It takes the union of current and new values.
	UpdateAttributes(ctx context.Context, ids *ttnpb.GatewayIdentifiers, current, new map[string]string) error
	// ValidateGatewayID validates the ID of the gateway.
	ValidateGatewayID(ctx context.Context, ids *ttnpb.GatewayIdentifiers) error
}
