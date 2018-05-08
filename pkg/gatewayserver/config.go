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

package gatewayserver

// Config represents the GatewayServer configuration.
type Config struct {
	NSTags []string `name:"network-servers.tags" description:"Network Server tags to accept to connect to"`

	DisableAuth bool `name:"disable-auth" description:"Disable gateway authentication, e.g. for debugging and testing purposes"`

	UDPAddress string `name:"udp-address" description:"Address for the UDP server to listen on"`
}
