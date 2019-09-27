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

	"go.thethings.network/lorawan-stack/pkg/types"
)

// LoRaAllianceTR005Draft2 is the LoRa Alliance defined format in Technical Recommendation TR005 Draft 2.
type LoRaAllianceTR005Draft2 struct {
	JoinEUI,
	DevEUI types.EUI64
	VendorID [2]byte
	ModelID  [2]byte
	DeviceValidationCode,
	SerialNumber,
	Proprietary string
}

// validTR005ExtensionChars defines the QR code alphanumeric character set except :, % and space.
const validTR005ExtensionChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ$*+-./"

func (m LoRaAllianceTR005Draft2) validateExtensionChars(s string) error {
	for _, r := range s {
		if strings.IndexRune(validTR005ExtensionChars, r) == -1 {
			return errCharacter.WithAttributes("r", r)
		}
	}
	return nil
}

// Validate implements the Data interface.
func (m LoRaAllianceTR005Draft2) Validate() error {
	for _, err := range []error{
		m.validateExtensionChars(m.DeviceValidationCode),
		m.validateExtensionChars(m.SerialNumber),
		m.validateExtensionChars(m.Proprietary),
	} {
		if err != nil {
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
		return errFormat
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
		return errFormat
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

// AuthenticatedEndDeviceIdentifiers implements the AuthenticatedEndDeviceIdentifiers interface.
func (m *LoRaAllianceTR005Draft2) AuthenticatedEndDeviceIdentifiers() (joinEUI, devEUI types.EUI64, authenticationCode string) {
	return m.JoinEUI, m.DevEUI, m.DeviceValidationCode
}
