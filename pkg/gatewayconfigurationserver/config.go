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

package gatewayconfigurationserver

import (
	bscups "go.thethings.network/lorawan-stack/v3/pkg/basicstation/cups"
	gcsv2 "go.thethings.network/lorawan-stack/v3/pkg/gatewayconfigurationserver/v2"
)

// Config contains the Gateway Configuration Server configuration.
type Config struct {
	// BasicStation defines the configuration for the BasicStation CUPS.
	BasicStation bscups.ServerConfig `name:"basic-station" description:"BasicStation CUPS configuration."`
	// TheThingsGateway defines the configuration for The Things Gateway CUPS.
	TheThingsGateway gcsv2.TheThingsGatewayConfig `name:"the-things-gateway" description:"The Things Gateway CUPS configuration."`
	// RequreAuth defines if the HTTP endpoints should require authentication or not.
	RequireAuth bool `name:"require-auth" description:"Require authentication for the HTTP endpoints."`
}
