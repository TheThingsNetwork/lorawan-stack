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

package translator

import (
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/udp"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// Translator defines the interface to convert messages from the ttn format to the packet
// forwarder network format.
type Translator interface {
	Downlink(*ttnpb.GatewayDown) (udp.Data, error)
	Upstream(message udp.Data, md Metadata) (*ttnpb.GatewayUp, error)
}

type Metadata struct {
	ID ttnpb.GatewayIdentifiers
	IP string

	Versions map[string]string
}

type translator struct {
	Logger log.Interface

	Location       *ttnpb.Location
	locationFromAS bool
}

// NewWithLocation returns a translator that inserts the given location in all messages
func NewWithLocation(logger log.Interface, location ttnpb.Location) Translator {
	return &translator{
		Logger: logger,

		Location:       &location,
		locationFromAS: true,
	}
}

// New returns a translator that converts Semtech UDP messages to protobuf messages
func New(logger log.Interface) Translator {
	return &translator{
		Logger: logger,
	}
}
