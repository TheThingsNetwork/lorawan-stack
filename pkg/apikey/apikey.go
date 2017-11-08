// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package apikey

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/random"
)

const (
	// Type is the JOSE type for the API Key.
	Type = "key"

	// alg is the JOSE algorithm for the API Key.
	alg = "secret"

	// entropy is the amount of entropy we use (in bytes).
	entropy = 32
)

var (
	// enc is the encoder we use.
	enc = base64.RawURLEncoding

	// header64 is the base64 encoded header.
	header64 string
)

// KeyType denotes the API key type.
type KeyType int

const (
	// TypeInvalid is an invalid type.
	TypeInvalid KeyType = iota

	// TypeApplication denotes it is an application API key.
	TypeApplication

	// TypeGateway denotes it is a gateway API key.
	TypeGateway
)

// String implements fmt.Stringer.
func (k KeyType) String() string {
	switch k {
	case TypeApplication:
		return "application"
	case TypeGateway:
		return "gateway"
	default:
		return "invalid type"
	}
}

type Header struct {
	Alg  string `json:"alg"`
	Type string `json:"typ"`
}

type Payload struct {
	Issuer string  `json:"iss"`
	Type   KeyType `json:"type"`
}

// GenerateApplicationAPIKey generates an API key with the following JOSE header:
// {
//    "typ": "key",
//    "alg": "secret",
// }
//
// a payload with the content:
// {
//    "iss": "<tenant>",
//    "type": "application",
// }
//
// and a JWS body that consists of random bytes.
func GenerateApplicationAPIKey(tenant string) (string, error) {
	return generateAPIKey(TypeApplication, tenant)
}

// GenerateGatewayAPIKey generates an API key with the following JOSE header:
// {
//    "typ": "key",
//    "alg": "secret",
// }
//
// a payload with the content:
// {
//    "iss": "<tenant>",
//    "type": "application",
// }
//
// and a JWS body that consists of random bytes.
func GenerateGatewayAPIKey(tenant string) (string, error) {
	return generateAPIKey(TypeGateway, tenant)
}

func generateAPIKey(typ KeyType, tenant string) (string, error) {
	payload, err := marshal(Payload{
		Issuer: tenant,
		Type:   typ,
	})
	if err != nil {
		return "", err
	}

	return header64 + "." + payload + "." + enc.EncodeToString(random.Bytes(entropy)), nil
}

// KeyPayload gets the payload from the base64 encoded key.
func KeyPayload(key string) (*Payload, error) {
	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("Invalid number of segments in key")
	}

	if len(parts[2]) <= 4 {
		return nil, fmt.Errorf("The API Key does not contain a valid secret")
	}

	head := new(Header)
	err := unmarshal([]byte(parts[0]), head)
	if err != nil {
		return nil, err
	}

	if head.Type != Type {
		return nil, fmt.Errorf("The received key is not an API Key")
	}

	if head.Alg != alg {
		return nil, fmt.Errorf("Unkown alg for API Key: %s", head.Alg)
	}

	payload := new(Payload)
	err = unmarshal([]byte(parts[1]), payload)
	if err != nil {
		return nil, err
	}

	if len(payload.Issuer) == 0 {
		return nil, fmt.Errorf("Invalid API Key: issuer is empty")
	}

	if payload.Type != TypeApplication && payload.Type != TypeGateway {
		return nil, fmt.Errorf("Invalid API Key: invalid type value")
	}

	return payload, nil
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
	header, err := marshal(Header{
		Type: Type,
		Alg:  alg,
	})
	if err != nil {
		panic(err)
	}

	header64 = header
}
