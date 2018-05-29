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

// MQTTConfig of the Gateway Server.
type MQTTConfig struct {
	Listen    string `name:"listen" description:"Address for the MQTT endpoint to listen on"`
	ListenTLS string `name:"listen-tls" description:"Address for the MQTTS endpoint to listen on"`
}

// Config represents the Gateway Server configuration.
type Config struct {
	NSTags []string `name:"network-servers.tags" description:"Network Server tags to accept to connect to"`

	DisableAuth bool `name:"disable-auth" description:"Disable gateway authentication, e.g. for debugging and testing purposes"`

	UDPAddress string `name:"udp.listen" description:"Address for the UDP endpoint to listen on"`

	MQTT MQTTConfig `name:"mqtt"`
}
