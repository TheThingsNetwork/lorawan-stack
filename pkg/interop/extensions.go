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

var tti = VendorID{0xec, 0x65, 0x6e}

// IsTheThingsIndustries returns true if the vendor ID is The Things Industries.
func (v VendorID) IsTheThingsIndustries() bool {
	return bytes.Equal(v[:], tti[:])
}

type TTIVendorID VendorID

// MarshalText returns the vendor ID of The Things Industries.
func (v TTIVendorID) MarshalText() ([]byte, error) {
	return tti.MarshalText()
}

// UnmarshalText returns an error if the vendor ID is not of The Things Industries.
func (v *TTIVendorID) UnmarshalText(data []byte) error {
	var vid VendorID
	if err := vid.UnmarshalText(data); err != nil {
		return err
	}
	if !vid.IsTheThingsIndustries() {
		return errInvalidVendorID.New()
	}
	*v = TTIVendorID(vid)
	return nil
}

// TTIHomeNSAns is HomeNSAns with vendor extension of The Things Industries.
type TTIHomeNSAns struct {
	HomeNSAns
	// HTenantID is the Tenant ID within a host HNetID.
	HTenantID string
	// HNSAddress is the Home Network Server address.
	HNSAddress string
}

func (m TTIHomeNSAns) MarshalJSON() ([]byte, error) {
	aux := struct {
		HomeNSAns
		VSExtension struct {
			VendorID TTIVendorID
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

func (m *TTIHomeNSAns) UnmarshalJSON(data []byte) error {
	var aux struct {
		HomeNSAns
		VSExtension struct {
			VendorID TTIVendorID
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
		HomeNSAns:  aux.HomeNSAns,
		HTenantID:  aux.VSExtension.Object.TTSV3.HTenantID,
		HNSAddress: aux.VSExtension.Object.TTSV3.HNSAddress,
	}
	return nil
}
