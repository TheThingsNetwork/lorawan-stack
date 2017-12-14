// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/random"
)

const (
	// alg is the JOSE algorithm for the Access Token and API Key.
	alg = "secret"

	// entropy is the amount of entropy we use (in bytes).
	entropy = 64
)

var (
	// enc is the encoder we use.
	enc = base64.RawURLEncoding
)

// Header is the JOSE header.
type Header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

// Payload is the payload used to generate API keys and Access Tokens.
type Payload struct {
	Issuer string     `json:"iss,omitempty"`
	Type   APIKeyType `json:"type,omitempty"`
}

func DecodeTokenOrKey(value string) (*Header, *Payload, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return nil, nil, errors.New("Invalid number of segments")
	}

	decHeader, err := enc.DecodeString(parts[0])
	if err != nil {
		return nil, nil, err
	}

	header := new(Header)
	if err := json.Unmarshal(decHeader, header); err != nil {
		return nil, nil, err
	}

	decPayload, err := enc.DecodeString(parts[1])
	if err != nil {
		return nil, nil, err
	}

	payload := new(Payload)
	if err := json.Unmarshal(decPayload, payload); err != nil {
		return nil, nil, err
	}

	return header, payload, nil
}

func generate(typ string, payload interface{}) (string, error) {
	encHeader, err := marshal(&Header{
		Algorithm: alg,
		Type:      typ,
	})
	if err != nil {
		return "", err
	}

	encPayload, err := marshal(payload)
	if err != nil {
		return "", err
	}

	return encHeader + "." + encPayload + "." + enc.EncodeToString(random.Bytes(entropy)), nil
}

func marshal(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return enc.EncodeToString(data), nil
}

func unmarshal(data []byte, v interface{}) error {
	js, err := enc.DecodeString(string(data))
	if err != nil {
		return err
	}

	return json.Unmarshal(js, v)
}
