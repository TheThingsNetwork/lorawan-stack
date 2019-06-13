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

package cups

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"math"

	"go.thethings.network/lorawan-stack/pkg/basicstation"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

var emptyClientCert = []byte{0x00, 0x00, 0x00, 0x00}

// UpdateInfoRequest is the contents of the update-info request.
type UpdateInfoRequest struct {
	Router             basicstation.EUI `json:"router"`
	CUPSURI            string           `json:"cupsUri"`
	LNSURI             string           `json:"tcUri"`
	CUPSCredentialsCRC uint32           `json:"cupsCredCrc"`
	LNSCredentialsCRC  uint32           `json:"tcCredCrc"`
	Station            string           `json:"station"`
	Model              string           `json:"model"`
	Package            string           `json:"package"`
	KeyCRCs            []uint32         `json:"keys"`
}

// UpdateInfoResponse is the response to the update-info request.
type UpdateInfoResponse struct {
	CUPSURI         string
	LNSURI          string
	CUPSCredentials []byte
	LNSCredentials  []byte
	SignatureKeyCRC uint32
	Signature       []byte
	UpdateData      []byte
}

// TLSCredentials appends the TLS trust certificate and client credentials.
// Only the leaf client certificate is included.
func TLSCredentials(trust *x509.Certificate, client *tls.Certificate) ([]byte, error) {
	var out []byte
	out = append(out, trust.Raw...)
	if client != nil {
		out = append(out, client.Certificate[0]...)
		switch privateKey := client.PrivateKey.(type) {
		case *rsa.PrivateKey:
			out = append(out, x509.MarshalPKCS1PrivateKey(privateKey)...)
		case *ecdsa.PrivateKey:
			privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
			if err != nil {
				return nil, err
			}
			out = append(out, privateKeyBytes...)
		default:
			return nil, errUnsupportedPrivateKey.WithAttributes("type", fmt.Sprintf("%T", client.PrivateKey))
		}
	}
	return out, nil
}

// TokenCredentials appends the TLS trust certificate and the contents of the Authorization header.
// Only the leaf of the trust certificate is considered.
func TokenCredentials(trust *x509.Certificate, authorization string) ([]byte, error) {
	var out []byte
	out = append(out, trust.Raw...)
	// TODO: Refactor when client side TLS is supported https://github.com/TheThingsNetwork/lorawan-stack/issues/137
	out = append(out, emptyClientCert...)
	out = append(out, []byte(fmt.Sprintf("%s%s%s", "Authorization: ", authorization, "\r\n"))...)
	return out, nil
}

var (
	errFieldLength           = errors.Define("field_length", "length of `{field}` (`{length}`) exceeds maximum `{maximum}`")
	errUnsupportedPrivateKey = errors.Define("unsupported_private_key", "the private key type `{type}` is not supported")
)

// MarshalBinary implements encoding.BinaryMarshaler.
func (r UpdateInfoResponse) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	lenBytes := make([]byte, 2)
	if uriLen := len(r.CUPSURI); uriLen <= math.MaxUint8 {
		b.WriteByte(uint8(uriLen)) // cupsUriLen
		b.WriteString(r.CUPSURI)   // cupsUri
	} else {
		return nil, errFieldLength.WithAttributes("field", "cupsUri", "length", uriLen, "maximum", math.MaxUint8)
	}
	if uriLen := len(r.LNSURI); uriLen <= math.MaxUint8 {
		b.WriteByte(uint8(uriLen)) // tcUriLen
		b.WriteString(r.LNSURI)    // tcUri
	} else {
		return nil, errFieldLength.WithAttributes("field", "tcUri", "length", uriLen, "maximum", math.MaxUint8)
	}
	if credLen := len(r.CUPSCredentials); credLen <= math.MaxUint16 {
		binary.LittleEndian.PutUint16(lenBytes, uint16(credLen))
		b.Write(lenBytes)          // cupsCredLen
		b.Write(r.CUPSCredentials) // cupsCred
	} else {
		return nil, errFieldLength.WithAttributes("field", "cupsCred", "length", credLen, "maximum", math.MaxUint16)
	}
	if credLen := len(r.LNSCredentials); credLen <= math.MaxUint16 {
		binary.LittleEndian.PutUint16(lenBytes, uint16(credLen))
		b.Write(lenBytes)         // tcCredLen
		b.Write(r.LNSCredentials) // tcCred
	} else {
		return nil, errFieldLength.WithAttributes("field", "tcCred", "length", credLen, "maximum", math.MaxUint16)
	}
	if sigLen := len(r.Signature); sigLen <= math.MaxUint16 {
		binary.LittleEndian.PutUint16(lenBytes, uint16(sigLen))
		b.Write(lenBytes) // sigLen
		crc := make([]byte, 4)
		binary.LittleEndian.PutUint32(crc, r.SignatureKeyCRC)
		b.Write(crc)         // sigCRC
		b.Write(r.Signature) // sig
	} else {
		return nil, errFieldLength.WithAttributes("field", "sig", "length", sigLen, "maximum", math.MaxUint16)
	}
	// NOTE: Please don't try sending 4GB updates on 32 bit systems. It will not work.
	if updLen := uint64(len(r.UpdateData)); updLen <= math.MaxUint32 {
		lenBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(lenBytes, uint32(updLen))
		b.Write(lenBytes)     // updLen
		b.Write(r.UpdateData) // updData
	} else {
		return nil, errFieldLength.WithAttributes("field", "updData", "length", updLen, "maximum", uint32(math.MaxUint32))
	}
	return b.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (r *UpdateInfoResponse) UnmarshalBinary(data []byte) error {
	b := bytes.NewBuffer(data)
	uriLen, err := b.ReadByte()
	if err != nil {
		return err
	}
	if uriLen > 0 {
		uri := make([]byte, uriLen)
		_, err = b.Read(uri)
		if err != nil {
			return err
		}
		r.CUPSURI = string(uri)
	} else {
		r.CUPSURI = ""
	}
	uriLen, err = b.ReadByte()
	if err != nil {
		return err
	}
	if uriLen > 0 {
		uri := make([]byte, uriLen)
		_, err = b.Read(uri)
		if err != nil {
			return err
		}
		r.LNSURI = string(uri)
	} else {
		r.LNSURI = ""
	}
	credLenBytes := make([]byte, 2)
	_, err = b.Read(credLenBytes)
	if err != nil {
		return err
	}
	credLen := binary.LittleEndian.Uint16(credLenBytes)
	if credLen > 0 {
		r.CUPSCredentials = make([]byte, credLen)
		_, err = b.Read(r.CUPSCredentials)
		if err != nil {
			return err
		}
	} else {
		r.CUPSCredentials = nil
	}
	_, err = b.Read(credLenBytes)
	if err != nil {
		return err
	}
	credLen = binary.LittleEndian.Uint16(credLenBytes)
	if credLen > 0 {
		r.LNSCredentials = make([]byte, credLen)
		_, err = b.Read(r.LNSCredentials)
		if err != nil {
			return err
		}
	} else {
		r.LNSCredentials = nil
	}
	sigLenBytes := make([]byte, 2)
	_, err = b.Read(sigLenBytes)
	if err != nil {
		return err
	}
	sigLen := binary.LittleEndian.Uint16(sigLenBytes)
	keyCRCBytes := make([]byte, 4)
	_, err = b.Read(keyCRCBytes)
	if err != nil {
		return err
	}
	r.SignatureKeyCRC = binary.LittleEndian.Uint32(keyCRCBytes)
	if sigLen > 0 {
		r.Signature = make([]byte, sigLen)
		_, err = b.Read(r.Signature)
		if err != nil {
			return err
		}
	} else {
		r.Signature = nil
	}
	updLenBytes := make([]byte, 4)
	_, err = b.Read(updLenBytes)
	if err != nil {
		return err
	}
	updLen := binary.LittleEndian.Uint32(updLenBytes)
	if updLen > 0 {
		r.UpdateData = make([]byte, updLen)
		_, err = b.Read(r.UpdateData)
		if err != nil {
			return err
		}
	} else {
		r.UpdateData = nil
	}
	return nil
}
