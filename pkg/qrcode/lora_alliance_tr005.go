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

package qrcode

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// LoRaAllianceTR005 is the LoRa Alliance defined format in Technical Recommendation TR005.
// See https://lora-alliance.org/wp-content/uploads/2020/11/TR005_LoRaWAN_Device_Identification_QR_Codes.pdf
type LoRaAllianceTR005 struct {
	JoinEUI,
	DevEUI types.EUI64
	VendorID,
	ModelID [2]byte
	Checksum,
	OwnerToken,
	SerialNumber,
	Proprietary string
}

// Encode implements the Data interface.
func (m *LoRaAllianceTR005) Encode(dev *ttnpb.EndDevice) error {
	if dev.JoinEui == nil {
		return errNoJoinEUI.New()
	}
	if dev.DevEui == nil {
		return errNoDevEUI.New()
	}
	*m = LoRaAllianceTR005{
		JoinEUI:    *dev.JoinEui,
		DevEUI:     *dev.DevEui,
		OwnerToken: dev.GetClaimAuthenticationCode().GetValue(),
	}
	return nil
}

func (LoRaAllianceTR005) validateExtensionChars(s string) error {
	if strings.IndexRune(s, ':') >= 0 {
		return errCharacter.WithAttributes("r", ":")
	}
	return nil
}

// Validate implements the Data interface.
func (m LoRaAllianceTR005) Validate() error {
	for _, ext := range []string{
		m.Checksum,
		m.OwnerToken,
		m.SerialNumber,
		m.Proprietary,
	} {
		if err := m.validateExtensionChars(ext); err != nil {
			return err
		}
	}
	return nil
}

// MarshalText implements the TextMarshaler interface.
func (m LoRaAllianceTR005) MarshalText() ([]byte, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	var ext string
	if m.Checksum != "" {
		ext += fmt.Sprintf(":C%s", m.Checksum)
	}
	if m.OwnerToken != "" {
		ext += fmt.Sprintf(":O%s", m.OwnerToken)
	}
	if m.SerialNumber != "" {
		ext += fmt.Sprintf(":S%s", m.SerialNumber)
	}
	if m.Proprietary != "" {
		ext += fmt.Sprintf(":P%s", m.Proprietary)
	}
	return []byte(fmt.Sprintf("LW:D0:%X:%X:%X%X%s", m.JoinEUI[:], m.DevEUI[:], m.VendorID[:], m.ModelID[:], ext)), nil
}

// UnmarshalText implements the TextUnmarshaler interface.
func (m *LoRaAllianceTR005) UnmarshalText(text []byte) error {
	parts := bytes.Split(text, []byte(":"))
	if len(parts) < 5 ||
		!bytes.Equal(parts[0], []byte("LW")) ||
		!bytes.Equal(parts[1], []byte("D0")) {
		return errFormat.New()
	}
	*m = LoRaAllianceTR005{}
	if err := m.JoinEUI.UnmarshalText(parts[2]); err != nil {
		return err
	}
	if err := m.DevEUI.UnmarshalText(parts[3]); err != nil {
		return err
	}
	prodID := make([]byte, hex.DecodedLen(len(parts[4])))
	if n, err := hex.Decode(prodID, parts[4]); err == nil && n == 4 {
		copy(m.VendorID[:], prodID[:2])
		copy(m.ModelID[:], prodID[2:])
	} else if n != 4 {
		return errFormat.New()
	} else {
		return err
	}
	if len(parts) > 5 {
		for _, ext := range parts[5:] {
			if len(ext) <= 1 {
				continue
			}
			ext := string(ext)
			val := ext[1:]
			switch ext[0] {
			case 'C':
				m.Checksum = val
			case 'O':
				m.OwnerToken = val
			case 'S':
				m.SerialNumber = val
			case 'P':
				m.Proprietary = val
			}
		}
	}
	return m.Validate()
}

// AuthenticatedEndDeviceIdentifiers implements the AuthenticatedEndDeviceIdentifiers interface.
func (m *LoRaAllianceTR005) AuthenticatedEndDeviceIdentifiers() (joinEUI, devEUI types.EUI64, authenticationCode string) {
	return m.JoinEUI, m.DevEUI, m.OwnerToken
}

type loRaAllianceTR005Format struct{}

func (loRaAllianceTR005Format) Format() *ttnpb.QRCodeFormat {
	return &ttnpb.QRCodeFormat{
		Name:        "LoRa Alliance TR005",
		Description: "Standard QR code format defined by LoRa Alliance.",
		FieldMask: &pbtypes.FieldMask{
			Paths: []string{
				"claim_authentication_code.value",
				"ids.dev_eui",
				"ids.join_eui",
			},
		},
	}
}

func (loRaAllianceTR005Format) New() EndDeviceData {
	return new(LoRaAllianceTR005)
}

func init() {
	RegisterEndDeviceFormat("tr005", new(loRaAllianceTR005Format))
}
