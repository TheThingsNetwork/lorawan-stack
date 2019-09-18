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

package upstream

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// Handler represents the upstream handler that connects to an upstream host.
type Handler interface {
	// GetHostName returns the hostname of the upstream.
	GetHostName() string
	// GetDevAddrPrefixes returns the device addr prefixes for this upstream handler. It's used to claim an uplink based on it's DevAddr.
	GetDevAddrPrefixes() []types.DevAddrPrefix
	// Setup performs all the preparation necessary to connect the handler to a particular upstream host.
	Setup() error
	// ConnectGateway informs the upstream handler that a particular gateway is connected to the front end.
	ConnectGateway(context.Context, ttnpb.GatewayIdentifiers, *io.Connection) error
	// HandleUp handles upstream messages.
	HandleUp(context.Context, ttnpb.GatewayIdentifiers, ttnpb.EndDeviceIdentifiers, *ttnpb.GatewayUp) error
}
