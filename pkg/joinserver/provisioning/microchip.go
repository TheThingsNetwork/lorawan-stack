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

package provisioning

import (
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/pkg/types"
	jose "gopkg.in/square/go-jose.v2"
)

var microchipPublicKeys = map[string][]byte{
	"B0uaDLLiyKe-SorP71F3BNoVrfY": []byte(`-----BEGIN CERTIFICATE-----
MIIByDCCAW6gAwIBAgIQdl97PjgCe2uPihhJX1FVXTAKBggqhkjOPQQDAjA9MSEw
HwYDVQQKDBhNaWNyb2NoaXAgVGVjaG5vbG9neSBJbmMxGDAWBgNVBAMMD0xvZyBT
aWduZXIgVGVzdDAeFw0xOTAxMTgyMDI5NDFaFw0xOTAyMTgyMDI5NDFaMD0xITAf
BgNVBAoMGE1pY3JvY2hpcCBUZWNobm9sb2d5IEluYzEYMBYGA1UEAwwPTG9nIFNp
Z25lciBUZXN0MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEHaTaoWgx1zG1JhnP
NbueEtfe926WJwbkHIyTBTB2aDzBUf3oRSFleYCJOEaRZlbEoQ4WiDDwCDeBd8GK
R70mN6NQME4wHQYDVR0OBBYEFAdLmgyy4sinvkqKz+9RdwTaFa32MB8GA1UdIwQY
MBaAFAdLmgyy4sinvkqKz+9RdwTaFa32MAwGA1UdEwEB/wQCMAAwCgYIKoZIzj0E
AwIDSAAwRQIhAI9jMSnc+HKKnjZ5ghmYVXYgPn9M9ae6gfE4AN5xekEZAiBNk7Pz
FVV78rUrxt7igKFg3mMLfE8Qeoh6dDKmRkbAEA==
-----END CERTIFICATE-----`),
	"7cCILlAOwYo1-PChGuoyUISMK3g": []byte(`-----BEGIN CERTIFICATE-----
MIIBxjCCAWygAwIBAgIQZGIWyMZI9cMcBZipXxTOWDAKBggqhkjOPQQDAjA8MSEw
HwYDVQQKDBhNaWNyb2NoaXAgVGVjaG5vbG9neSBJbmMxFzAVBgNVBAMMDkxvZyBT
aWduZXIgMDAxMB4XDTE5MDEyMjAwMjc0MloXDTE5MDcyMjAwMjc0MlowPDEhMB8G
A1UECgwYTWljcm9jaGlwIFRlY2hub2xvZ3kgSW5jMRcwFQYDVQQDDA5Mb2cgU2ln
bmVyIDAwMTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABEu8/ZyRdTu4N0kuu76C
R1JR5vz04EuRqL4TQxMinRiUc3Htqy38O6HrXo2qmNoyrO0xd2I2pfQhXWYuLT35
MGWjUDBOMB0GA1UdDgQWBBTtwIguUA7BijX48KEa6jJQhIwreDAfBgNVHSMEGDAW
gBTtwIguUA7BijX48KEa6jJQhIwreDAMBgNVHRMBAf8EAjAAMAoGCCqGSM49BAMC
A0gAMEUCIQD9/x9zxmHkeWGwjEq67QsQqBVmoY8k6PvFVr4Bz1tYOwIgYfck+fv/
pno8+2vVTkQDhcinNrgoPLQORzV5/l/b4z4=
-----END CERTIFICATE-----`),
}

// microchip is a Microchip device provisioner.
type microchip struct {
	keys map[string]interface{}
}

type microchipEntry struct {
	jose.JSONWebSignature
}

func (m *microchipEntry) UnmarshalJSON(data []byte) error {
	jws, err := jose.ParseSigned(string(data))
	if err != nil {
		return err
	}
	*m = microchipEntry{JSONWebSignature: *jws}
	return nil
}

var errMicrochipPublicKey = errors.DefineInvalidArgument("microchip_public_key", "unknown Microchip public key ID `{id}`")

// Decode decodes the given data and returns a struct.
func (p *microchip) Decode(data []byte) ([]*pbtypes.Struct, error) {
	entries := make([]microchipEntry, 0)
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	res := make([]*pbtypes.Struct, 0, len(entries))
	for _, jws := range entries {
		kid := jws.Signatures[0].Header.KeyID
		key, ok := p.keys[kid]
		if !ok {
			return nil, errMicrochipPublicKey.WithAttributes("id", kid)
		}
		buf, err := jws.Verify(key)
		if err != nil {
			return nil, err
		}
		m := make(map[string]interface{})
		if err := json.Unmarshal(buf, &m); err != nil {
			return nil, err
		}
		s, err := gogoproto.Struct(m)
		if err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return res, nil
}

// DefaultJoinEUI returns the default JoinEUI 70B3D57ED0000000.
func (p *microchip) DefaultJoinEUI(entry *pbtypes.Struct) (types.EUI64, error) {
	return types.EUI64{0x70, 0xB3, 0xD5, 0x7E, 0xD0, 0x00, 0x00, 0x00}, nil
}

// DefaultDevEUI returns the first 8 bytes of the serial number as DevEUI.
func (p *microchip) DefaultDevEUI(entry *pbtypes.Struct) (types.EUI64, error) {
	sn, err := hex.DecodeString(entry.Fields["uniqueId"].GetStringValue())
	if err != nil {
		return types.EUI64{}, errEntry.WithCause(err)
	}
	var eui types.EUI64
	copy(eui[:], sn[:8])
	return eui, nil
}

// DeviceID returns the device ID formatted as sn-{uniqueId}.
func (p *microchip) DeviceID(joinEUI, devEUI types.EUI64, entry *pbtypes.Struct) (string, error) {
	sn, err := p.UniqueID(entry)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sn-%s", strings.ToLower(sn)), nil
}

// UniqueID returns the serial number.
func (p *microchip) UniqueID(entry *pbtypes.Struct) (string, error) {
	sn := entry.Fields["uniqueId"].GetStringValue()
	if sn == "" {
		return "", errEntry
	}
	return strings.ToUpper(sn), nil
}

func init() {
	keys := make(map[string]interface{}, len(microchipPublicKeys))
	for kid, key := range microchipPublicKeys {
		block, _ := pem.Decode(key)
		if block == nil {
			panic(fmt.Sprintf("invalid Microchip public key %v", kid))
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			panic(err)
		}
		keys[kid] = cert.PublicKey
	}

	Register("microchip", &microchip{
		keys: keys,
	})
}
