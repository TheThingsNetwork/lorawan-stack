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

import "go.thethings.network/lorawan-stack/pkg/crypto"

// LinkMode defines how applications are linked to their Network Server.
type LinkMode int

const (
	// LinkAll links all applications in the link registry to their Network Server automatically.
	LinkAll LinkMode = iota
	// LinkExplicit links applications on request.
	LinkExplicit
)

// Config represents the ApplicationServer configuration.
type Config struct {
	LinkMode LinkMode `name:"link-mode" description:"Mode used to link applications to their Network Server"`
	Devices  DeviceRegistry
	Links    LinkRegistry
	KeyVault crypto.KeyVault
}
