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

const formatIDLoRaAllianceTR005Draft2 = "tr005draft2"

// LoRaAllianceTR005Draft2 is the LoRa Alliance defined format in Technical Recommendation TR005 Draft 2.
type LoRaAllianceTR005Draft2 struct {
	JoinEUI,
	DevEUI types.EUI64
	VendorID,
	ModelID [2]byte
	DeviceValidationCode,
	SerialNumber,
	Proprietary string
}

// Encode implements the Data interface.
func (m *LoRaAllianceTR005Draft2) Encode(dev *ttnpb.EndDevice) error {
	if dev.Ids == nil || dev.Ids.JoinEui == nil {
		return errNoJoinEUI.New()
	}
	if dev.Ids.DevEui == nil {
		return errNoDevEUI.New()
	}
	*m = LoRaAllianceTR005Draft2{
		JoinEUI:              types.MustEUI64(dev.Ids.JoinEui).OrZero(),
		DevEUI:               types.MustEUI64(dev.Ids.DevEui).OrZero(),
		DeviceValidationCode: dev.GetClaimAuthenticationCode().GetValue(),
	}
	return nil
}

// validTR005Draft2ExtensionChars defines the QR code alphanumeric character set except :, % and space.
const validTR005Draft2ExtensionChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ$*+-./"

func (LoRaAllianceTR005Draft2) validateExtensionChars(s string) error {
	for _, r := range s {
		if strings.IndexRune(validTR005Draft2ExtensionChars, r) == -1 {
			return errCharacter.WithAttributes("r", r)
		}
	}
	return nil
}

// Validate implements the Data interface.
func (m LoRaAllianceTR005Draft2) Validate() error {
	for _, ext := range []string{
		m.DeviceValidationCode,
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
func (m LoRaAllianceTR005Draft2) MarshalText() ([]byte, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	var ext string
	if m.DeviceValidationCode != "" {
		ext += fmt.Sprintf("%%V%s", m.DeviceValidationCode)
	}
	if m.SerialNumber != "" {
		ext += fmt.Sprintf("%%S%s", m.SerialNumber)
	}
	if m.Proprietary != "" {
		ext += fmt.Sprintf("%%P%s", m.Proprietary)
	}
	if ext != "" {
		ext = ":" + ext
	}
	return []byte(fmt.Sprintf("URN:LW:DP:%X:%X:%X%X%s", m.JoinEUI[:], m.DevEUI[:], m.VendorID[:], m.ModelID[:], ext)), nil
}

// UnmarshalText implements the TextUnmarshaler interface.
func (m *LoRaAllianceTR005Draft2) UnmarshalText(text []byte) error {
	parts := bytes.SplitN(text, []byte(":"), 7)
	if len(parts) < 6 ||
		!bytes.Equal(parts[0], []byte("URN")) ||
		!bytes.Equal(parts[1], []byte("LW")) ||
		!bytes.Equal(parts[2], []byte("DP")) {
		return errFormat.New()
	}
	*m = LoRaAllianceTR005Draft2{}
	if err := m.JoinEUI.UnmarshalText(parts[3]); err != nil {
		return err
	}
	if err := m.DevEUI.UnmarshalText(parts[4]); err != nil {
		return err
	}
	prodID := make([]byte, hex.DecodedLen(len(parts[5])))
	if n, err := hex.Decode(prodID, parts[5]); err == nil && n == 4 {
		copy(m.VendorID[:], prodID[:2])
		copy(m.ModelID[:], prodID[2:])
	} else if n != 4 {
		return errFormat.New()
	} else {
		return err
	}
	if len(parts) == 7 {
		exts := strings.ReplaceAll(string(parts[6]), "%25", "%")
		for _, ext := range strings.Split(exts, "%") {
			if len(ext) < 1 {
				continue
			}
			val := ext[1:]
			switch ext[0] {
			case 'V':
				m.DeviceValidationCode = val
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
func (m *LoRaAllianceTR005Draft2) FormatID() string {
	return formatIDLoRaAllianceTR005Draft2
}

// EndDeviceTemplate implements the Data interface.
func (m *LoRaAllianceTR005Draft2) EndDeviceTemplate() *ttnpb.EndDeviceTemplate {
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
				Value: m.DeviceValidationCode,
			},
			VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
				VendorId:        uint32(vendorID),
				VendorProfileId: uint32(vendorProfileID),
				SerialNumber:    m.SerialNumber,
			},
		},
		FieldMask: ttnpb.FieldMask(paths...),
	}
}

// LoRaAllianceTR005Draft2Format implements the LoRa Alliance TR005 Draft2 Format.
type LoRaAllianceTR005Draft2Format struct{}

// Format implements the Format interface.
func (LoRaAllianceTR005Draft2Format) Format() *ttnpb.QRCodeFormat {
	return &ttnpb.QRCodeFormat{
		Name:        "LoRa Alliance TR005 Draft 2 (DEPRECATED)",
		Description: "Standard QR code format defined by LoRa Alliance.",
		FieldMask: ttnpb.FieldMask(
			"claim_authentication_code.value",
			"ids.dev_eui",
			"ids.join_eui",
		),
	}
}

// ID is the identifier of the format as a string.
func (LoRaAllianceTR005Draft2Format) ID() string {
	return formatIDLoRaAllianceTR005Draft2
}

// New implements the Format interface.
func (LoRaAllianceTR005Draft2Format) New() Data {
	return new(LoRaAllianceTR005Draft2)
}
