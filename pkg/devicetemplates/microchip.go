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

package devicetemplates

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"strings"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/pkg/provisioning"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
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
	"8VeKGdyU2d8wev6_VzNJOBOv-cA": []byte(`-----BEGIN CERTIFICATE-----
MIIBxzCCAWygAwIBAgIQc6HyMjrkT2TOY17FAX7XVjAKBggqhkjOPQQDAjA8MSEw
HwYDVQQKDBhNaWNyb2NoaXAgVGVjaG5vbG9neSBJbmMxFzAVBgNVBAMMDkxvZyBT
aWduZXIgMDAyMB4XDTE5MDgxNTE5NDc1OVoXDTIwMDgxNTE5NDc1OVowPDEhMB8G
A1UECgwYTWljcm9jaGlwIFRlY2hub2xvZ3kgSW5jMRcwFQYDVQQDDA5Mb2cgU2ln
bmVyIDAwMjBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABLCLrgPlT3OezntD9lC2
ShwUhlx07fiq/VETJ+ITUAwbgrPjB/Xi9GchLIM7FwZSUGOEqRA6KtH32XMpTGHK
mCCjUDBOMB0GA1UdDgQWBBTxV4oZ3JTZ3zB6/r9XM0k4E6/5wDAfBgNVHSMEGDAW
gBTxV4oZ3JTZ3zB6/r9XM0k4E6/5wDAMBgNVHRMBAf8EAjAAMAoGCCqGSM49BAMC
A0kAMEYCIQDKHgctLnq/zNqfB+1v0KRhDVPvRf6Dimt8aW9WLS0NWAIhAJvUe3uJ
pkMG4zpov9FCoj4G340idEadm7mVbAd5GOB9
-----END CERTIFICATE-----`),
}

var joinEUI = types.EUI64{0x70, 0xb3, 0xd5, 0x7e, 0xd0, 0x00, 0x00, 0x00}

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

var (
	errMicrochipData      = errors.DefineInvalidArgument("microchip_data", "invalid Microchip data")
	errMicrochipPublicKey = errors.DefineInvalidArgument("microchip_public_key", "unknown Microchip public key ID `{id}`")
)

// microchipATECC608AMAHTNT is a Microchip ATECC608A-MAHTN-T device provisioner.
type microchipATECC608AMAHTNT struct {
	keys map[string]interface{}
}

func (m *microchipATECC608AMAHTNT) Format() *ttnpb.EndDeviceTemplateFormat {
	return &ttnpb.EndDeviceTemplateFormat{
		Name:           "Microchip ATECC608A-MAHTN-T Manifest File",
		Description:    "JSON manifest file received through Microchip Purchasing & Client Services.",
		FileExtensions: []string{".json"},
	}
}

// Convert decodes the given manifest data.
// The input data is an array of JWS (JSON Web Signatures).
func (m *microchipATECC608AMAHTNT) Convert(ctx context.Context, r io.Reader, ch chan<- *ttnpb.EndDeviceTemplate) error {
	defer close(ch)

	dec := json.NewDecoder(r)
	delim, err := dec.Token()
	if err != nil {
		return errMicrochipData.WithCause(err)
	}
	if _, ok := delim.(json.Delim); !ok {
		return errMicrochipData
	}

	for dec.More() {
		var jws microchipEntry
		if err := dec.Decode(&jws); err != nil {
			return errMicrochipData.WithCause(err)
		}
		kid := jws.Signatures[0].Header.KeyID
		key, ok := m.keys[kid]
		if !ok {
			return errMicrochipPublicKey.WithAttributes("id", kid)
		}
		buf, err := jws.Verify(key)
		if err != nil {
			return errMicrochipData.WithCause(err)
		}
		m := make(map[string]interface{})
		if err := json.Unmarshal(buf, &m); err != nil {
			return errMicrochipData.WithCause(err)
		}
		s, err := gogoproto.Struct(m)
		if err != nil {
			return errMicrochipData.WithCause(err)
		}
		ch <- &ttnpb.EndDeviceTemplate{
			EndDevice: ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					JoinEUI: &joinEUI,
				},
				ProvisionerID:    provisioning.Microchip,
				ProvisioningData: s,
			},
			FieldMask: pbtypes.FieldMask{
				Paths: []string{
					"ids.join_eui",
					"provisioner_id",
					"provisioning_data",
				},
			},
			MappingKey: s.Fields["uniqueId"].GetStringValue(),
		}
	}
	return nil
}

// microchipATECC608TNGLORA is a Microchip ATECC608A-TNGLORA device provisioner.
type microchipATECC608TNGLORA struct {
	keys map[string]interface{}
}

func (m *microchipATECC608TNGLORA) Format() *ttnpb.EndDeviceTemplateFormat {
	return &ttnpb.EndDeviceTemplateFormat{
		Name:           "Microchip ATECC608A-TNGLORA Manifest File",
		Description:    "JSON manifest file received through Microchip Purchasing & Client Services.",
		FileExtensions: []string{".json"},
	}
}

// Convert decodes the given manifest data.
// The input data is an array of JWS (JSON Web Signatures).
func (m *microchipATECC608TNGLORA) Convert(ctx context.Context, r io.Reader, ch chan<- *ttnpb.EndDeviceTemplate) error {
	defer close(ch)

	dec := json.NewDecoder(r)
	delim, err := dec.Token()
	if err != nil {
		return errMicrochipData.WithCause(err)
	}
	if _, ok := delim.(json.Delim); !ok {
		return errMicrochipData
	}

	for dec.More() {
		var jws microchipEntry
		if err := dec.Decode(&jws); err != nil {
			return errMicrochipData.WithCause(err)
		}
		kid := jws.Signatures[0].Header.KeyID
		key, ok := m.keys[kid]
		if !ok {
			return errMicrochipPublicKey.WithAttributes("id", kid)
		}
		buf, err := jws.Verify(key)
		if err != nil {
			return errMicrochipData.WithCause(err)
		}
		// publicKeySet.keys[0].x5c[0] contains the base64 certificate of the secure element.
		// The second value in the CN is the hex encoded IEEE issued DevEUI.
		data := struct {
			PublicKeySet struct {
				Keys []struct {
					X5C []string `json:"x5c"`
				} `json:"keys"`
			} `json:"publicKeySet"`
		}{}
		if err := json.Unmarshal(buf, &data); err != nil {
			return errMicrochipData.WithCause(err)
		}
		if len(data.PublicKeySet.Keys) < 1 || len(data.PublicKeySet.Keys[0].X5C) < 1 {
			return errMicrochipData
		}
		certBuf, err := base64.StdEncoding.DecodeString(data.PublicKeySet.Keys[0].X5C[0])
		if err != nil {
			return errMicrochipData.WithCause(err)
		}
		cert, err := x509.ParseCertificate(certBuf)
		if err != nil {
			return errMicrochipData.WithCause(err)
		}
		cnParts := strings.SplitN(cert.Subject.CommonName, " ", 3)
		if len(cnParts) != 3 {
			return errMicrochipData.WithCause(err)
		}
		var devEUI types.EUI64
		if err := devEUI.UnmarshalText([]byte(cnParts[1])); err != nil {
			return errMicrochipData.WithCause(err)
		}
		m := make(map[string]interface{})
		if err := json.Unmarshal(buf, &m); err != nil {
			return errMicrochipData.WithCause(err)
		}
		s, err := gogoproto.Struct(m)
		if err != nil {
			return errMicrochipData.WithCause(err)
		}
		ch <- &ttnpb.EndDeviceTemplate{
			EndDevice: ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID: strings.ToLower(fmt.Sprintf("eui-%s", devEUI)),
					JoinEUI:  &joinEUI,
					DevEUI:   &devEUI,
				},
				ProvisionerID:    provisioning.Microchip,
				ProvisioningData: s,
			},
			FieldMask: pbtypes.FieldMask{
				Paths: []string{
					"ids.device_id",
					"ids.dev_eui",
					"ids.join_eui",
					"provisioner_id",
					"provisioning_data",
				},
			},
			MappingKey: s.Fields["uniqueId"].GetStringValue(),
		}
	}
	return nil
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

	RegisterConverter("microchip-atecc608a-mahtn-t", &microchipATECC608AMAHTNT{
		keys: keys,
	})
	RegisterConverter("microchip-atecc608a-tnglora", &microchipATECC608TNGLORA{
		keys: keys,
	})
}
