// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package apikey

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/random"
)

var (
	// enc is the encoder we use
	enc = base64.RawURLEncoding

	// typ is the JOSE type for the API Key
	typ = "key"

	// alg is the JOSE algorithm for the API Key
	alg = "secret"

	// entropy is the amount of entropy we use (in bytes)
	entropy = 32

	// header64 is the base64 encoded header
	header64 string
)

type header struct {
	Alg  string `json:"alg"`
	Type string `json:"typ"`
}

type payload struct {
	Issuer string `json:"iss"`
}

// GenerateAPIKey generates an API key with the JOSE header:
// {"typ":"key", "iss": "<tenant>"} and a JWS body that consists of random bytes.
func GenerateAPIKey(tenant string) (string, error) {
	payload, err := marshal(payload{
		Issuer: tenant,
	})
	if err != nil {
		return "", err
	}

	return header64 + "." + payload + "." + enc.EncodeToString(random.Bytes(entropy)), nil
}

// KeyTenant gets the tenant from the base64 encoded key.
func KeyTenant(key string) (string, error) {
	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("Invalid number of segments in key")
	}

	if len(parts[2]) <= 4 {
		return "", fmt.Errorf("The API Key does not contain a valid secret")
	}

	head := new(header)
	err := unmarshal([]byte(parts[0]), head)
	if err != nil {
		return "", err
	}

	if head.Type != typ {
		return "", fmt.Errorf("The received key is not an API Key")
	}

	if head.Alg != alg {
		return "", fmt.Errorf("Unkown alg for API Key: %s", head.Alg)
	}

	payload := new(payload)
	err = unmarshal([]byte(parts[1]), payload)
	if err != nil {
		return "", err
	}

	if payload.Issuer == "" {
		return "", fmt.Errorf("The API Key does not contain an issuer")
	}

	return payload.Issuer, nil
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

func init() {
	header, err := marshal(header{
		Type: typ,
		Alg:  alg,
	})
	if err != nil {
		panic(err)
	}

	header64 = header
}
