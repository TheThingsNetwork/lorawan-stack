// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// Header is a JOSE header.
type Header struct {
	Type      string `json:"typ"`
	Algorithm string `json:"alg"`
}

// JOSEHeader returns the decoded JOSE header claims from a JOSE-compliant token.
func JOSEHeader(key string) (*Header, error) {
	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("Invalid number of segments")
	}

	data, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}

	header := new(Header)
	if err := json.Unmarshal(data, header); err != nil {
		return nil, err
	}

	return header, nil
}
