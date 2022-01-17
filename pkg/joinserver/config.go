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

package joinserver

import "go.thethings.network/lorawan-stack/v3/pkg/types"

// Config represents the JoinServer configuration.
type Config struct {
	Devices                       DeviceRegistry                       `name:"-"`
	Keys                          KeyRegistry                          `name:"-"`
	ApplicationActivationSettings ApplicationActivationSettingRegistry `name:"-"`
	JoinEUIPrefixes               []types.EUI64Prefix                  `name:"join-eui-prefix" description:"JoinEUI prefixes handled by this Join Server"`
	DefaultJoinEUI                types.EUI64                          `name:"default-join-eui" description:"Default JoinEUI for this Join Server"`
	DeviceKEKLabel                string                               `name:"device-kek-label" description:"Label of KEK used to encrypt device keys at rest"`
	DevNonceLimit                 int                                  `name:"dev-nonce-limit" description:"Amount of DevNonces stored per device"`
}
