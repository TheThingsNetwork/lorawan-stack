// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package io

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"

	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
)

// Decoder is the interface for the functionality that reads and decodes entities
// from an io.Reader, typically os.Stdin.
type Decoder interface {
	Decode(data interface{}) (paths []string, err error)
}

type jsonDecoder struct {
	rd  *bufio.Reader
	dec *json.Decoder
}

// NewJSONDecoder returns a new Decoder on top of r, and that uses the common JSON
// format used in The Things Stack.
func NewJSONDecoder(r io.Reader) Decoder {
	rd := bufio.NewReader(r)
	return &jsonDecoder{
		rd:  rd,
		dec: json.NewDecoder(rd),
	}
}

func (r *jsonDecoder) Decode(data interface{}) (paths []string, err error) {
	t, err := r.rd.ReadByte()
	if err != nil {
		return nil, err
	}
	if t == '{' {
		if err := r.rd.UnreadByte(); err != nil {
			return nil, err
		}
	}
	var obj json.RawMessage
	if err = r.dec.Decode(&obj); err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err = json.Unmarshal(obj, &m); err != nil {
		return nil, err
	}
	paths = fieldPaths(m, "")
	b := bytes.NewBuffer(obj)
	if err = jsonpb.TTN().NewDecoder(b).Decode(data); err != nil {
		return nil, err
	}
	r.rd = bufio.NewReader(io.MultiReader(r.dec.Buffered(), r.rd))
	r.dec = json.NewDecoder(r.rd)
	return paths, nil
}

func fieldPaths(m map[string]interface{}, prefix string) (paths []string) {
	for path, sub := range m {
		if m, ok := sub.(map[string]interface{}); ok {
			paths = append(paths, fieldPaths(m, prefix+path+".")...)
		} else {
			paths = append(paths, prefix+path)
		}
	}
	return paths
}
