// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package interop

import (
	"bytes"
	"encoding/json"
)

// TTIVendorID is the vendor ID of The Things Industries.
var TTIVendorID = VendorID{0xec, 0x65, 0x6e}

// TTIVendorIDType is the custom type for The Things Industries vendor ID.
type TTIVendorIDType VendorID

// MarshalText returns the vendor ID of The Things Industries.
func (TTIVendorIDType) MarshalText() ([]byte, error) {
	return TTIVendorID.MarshalText()
}

// UnmarshalText returns an error if the vendor ID is not of The Things Industries.
func (v *TTIVendorIDType) UnmarshalText(data []byte) error {
	var vid VendorID
	if err := vid.UnmarshalText(data); err != nil {
		return err
	}
	if !bytes.Equal(vid[:], TTIVendorID[:]) {
		return errInvalidVendorID.New()
	}
	*v = TTIVendorIDType(vid)
	return nil
}

// TTIVSExtension is vendor extension of The Things Industries.
type TTIVSExtension struct {
	// HTenantID is the Tenant ID within a host HNetID.
	HTenantID string
	// HNSAddress is the Home Network Server address.
	HNSAddress string
}

// TTIHomeNSAns is HomeNSAns with vendor extension of The Things Industries.
type TTIHomeNSAns struct {
	HomeNSAns
	TTIVSExtension
}

// MarshalJSON implements json.Unmarshaler.
func (m TTIHomeNSAns) MarshalJSON() ([]byte, error) {
	aux := struct {
		HomeNSAns
		VSExtension struct {
			VendorID TTIVendorIDType
			Object   struct {
				TTSV3 struct {
					HTenantID  string `json:",omitempty"`
					HNSAddress string `json:",omitempty"`
				}
			}
		}
	}{
		HomeNSAns: m.HomeNSAns,
	}
	aux.VSExtension.Object.TTSV3.HTenantID = m.HTenantID
	aux.VSExtension.Object.TTSV3.HNSAddress = m.HNSAddress
	return json.Marshal(aux)
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *TTIHomeNSAns) UnmarshalJSON(data []byte) error {
	var aux struct {
		HomeNSAns
		VSExtension struct {
			VendorID TTIVendorIDType
			Object   struct {
				TTSV3 struct {
					HTenantID  string `json:",omitempty"`
					HNSAddress string `json:",omitempty"`
				}
			}
		}
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*m = TTIHomeNSAns{
		HomeNSAns: aux.HomeNSAns,
		TTIVSExtension: TTIVSExtension{
			HTenantID:  aux.VSExtension.Object.TTSV3.HTenantID,
			HNSAddress: aux.VSExtension.Object.TTSV3.HNSAddress,
		},
	}
	return nil
}
