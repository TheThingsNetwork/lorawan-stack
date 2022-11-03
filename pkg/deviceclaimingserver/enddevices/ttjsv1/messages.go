// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package ttjsv1

import (
	"encoding/json"
	"fmt"
	"strconv"

	"go.thethings.network/lorawan-stack/v3/pkg/interop"
)

// VendorSpecific defines vendor specific fields.
type VendorSpecific struct {
	OUI  OUI                    `json:"oui"`
	Data interop.TTIVSExtension `json:"data"`
}

// MarshalJSON implements json.Marshaler.
func (vs VendorSpecific) MarshalJSON() ([]byte, error) {
	str := struct {
		OUI  OUI `json:"oui"`
		Data struct {
			TTSV3 struct {
				HTenantID  string `json:",omitempty"`
				HNSAddress string `json:",omitempty"`
			}
		} `json:"data"`
	}{
		OUI: vs.OUI,
	}
	str.Data.TTSV3.HTenantID = vs.Data.HTenantID
	str.Data.TTSV3.HNSAddress = vs.Data.HNSAddress
	return json.Marshal(str)
}

// UnmarshalJSON implements json.Unmarshaler.
func (vs *VendorSpecific) UnmarshalJSON(data []byte) error {
	var str struct {
		OUI  OUI `json:"oui"`
		Data struct {
			TTSV3 struct {
				HTenantID  string `json:",omitempty"`
				HNSAddress string `json:",omitempty"`
			}
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*vs = VendorSpecific{
		OUI: str.OUI,
		Data: interop.TTIVSExtension{
			HTenantID:  str.Data.TTSV3.HTenantID,
			HNSAddress: str.Data.TTSV3.HNSAddress,
		},
	}
	return nil
}

type claimData struct {
	HomeNetID      string         `json:"homeNetID"`
	HomeNSID       string         `json:"homeNSID"`
	VendorSpecific VendorSpecific `json:"vendorSpecific"`
}

// OUI is the Organisation Unique Identifier.
type OUI uint32

// MarshalText implements encoding.TextUnmarshaler interface.
// This makes sure that the value sent to The Things Join Server is six upper case (UTF-8) hex characters.
func (oui OUI) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%06X", oui)), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (oui *OUI) UnmarshalText(data []byte) error {
	raw, err := strconv.ParseInt(string(data), 16, 32)
	if err != nil {
		return err
	}
	*oui = OUI(raw)
	return nil
}

type claimRequest struct {
	claimData
	OwnerToken string `json:"ownerToken"`
	Locked     bool   `json:"locked"`
}

// errorResponse is a message that may be returned by The Things Join Server in case of an error.
type errorResponse struct {
	Message string `json:"message"`
}
