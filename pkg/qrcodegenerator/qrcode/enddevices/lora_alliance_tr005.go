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

package enddevices

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

const (
	formatIDLoRaAllianceTR005 = "tr005"
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
	if dev.Ids == nil || dev.Ids.JoinEui == nil {
		return errNoJoinEUI.New()
	}
	if dev.Ids.DevEui == nil {
		return errNoDevEUI.New()
	}
	*m = LoRaAllianceTR005{
		JoinEUI:    types.MustEUI64(dev.Ids.JoinEui).OrZero(),
		DevEUI:     types.MustEUI64(dev.Ids.DevEui).OrZero(),
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

// FormatID implements the Data interface.
func (m *LoRaAllianceTR005) FormatID() string {
	return formatIDLoRaAllianceTR005
}

// EndDeviceTemplate implements the Data interface.
func (m *LoRaAllianceTR005) EndDeviceTemplate() *ttnpb.EndDeviceTemplate {
	paths := []string{
		"ids",
		"claim_authentication_code",
	}
	var vendorID, vendorProfileID uint16
	if m.VendorID != [2]byte{} {
		vendorID = binary.BigEndian.Uint16(m.VendorID[:])
	}
	if m.ModelID != [2]byte{} {
		vendorProfileID = binary.BigEndian.Uint16(m.ModelID[:])
	}
	return &ttnpb.EndDeviceTemplate{
		EndDevice: &ttnpb.EndDevice{
			Ids: &ttnpb.EndDeviceIdentifiers{
				DevEui:  m.DevEUI.Bytes(),
				JoinEui: m.JoinEUI.Bytes(),
			},
			ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
				Value: m.OwnerToken,
			},
			VendorId:        uint32(vendorID),
			VendorProfileId: uint32(vendorProfileID),
			SerialNumber:    m.SerialNumber,
		},
		FieldMask: ttnpb.FieldMask(paths...),
	}
}

// LoRaAllianceTR005Format implements the LoRa Alliance TR005 Format.
type LoRaAllianceTR005Format struct{}

// Format implements the Format interface.
func (LoRaAllianceTR005Format) Format() *ttnpb.QRCodeFormat {
	return &ttnpb.QRCodeFormat{
		Name:        "LoRa Alliance TR005",
		Description: "Standard QR code format defined by LoRa Alliance.",
		FieldMask: ttnpb.FieldMask(
			"claim_authentication_code.value",
			"ids.dev_eui",
			"ids.join_eui",
		),
	}
}

// ID is the identifier of the format as a string.
func (LoRaAllianceTR005Format) ID() string {
	return formatIDLoRaAllianceTR005
}

// New implements the Format interface.
func (LoRaAllianceTR005Format) New() Data {
	return new(LoRaAllianceTR005)
}
