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
	"encoding/base64"
	"encoding/hex"
	"io"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
)

// Decoder is the interface for the functionality that reads and decodes entities
// from an io.Reader, typically os.Stdin.
type Decoder interface {
	Decode(msg any) error
}

type jsonDecoder struct {
	inArray bool
	it      *jsoniter.Iterator
}

// NewJSONDecoder returns a new Decoder on top of r, and that uses the common JSON
// format used in The Things Stack.
func NewJSONDecoder(r io.Reader) Decoder {
	return &jsonDecoder{
		it: jsoniter.Parse(jsoniter.ConfigCompatibleWithStandardLibrary, r, 1024),
	}
}

var errJSONToken = errors.DefineInvalidArgument("json_token", "invalid JSON token")

func readJSONNext(it *jsoniter.Iterator) (jsoniter.ValueType, error) {
	next := it.WhatIsNext()
	if err := it.Error; err != nil {
		return jsoniter.InvalidValue, errJSONToken.WithCause(err)
	}
	return next, nil
}

func readJSONValue(it *jsoniter.Iterator, msg any) error {
	var buffer jsoniter.RawMessage
	it.ReadVal(&buffer)
	if err := it.Error; err != nil {
		return errJSONToken.WithCause(err)
	}
	return jsonpb.TTN().Unmarshal(buffer, msg)
}

func readJSONArray(it *jsoniter.Iterator) error {
	_ = it.ReadArray()
	if err := it.Error; err != nil {
		return errJSONToken.WithCause(err)
	}
	return nil
}

func (r *jsonDecoder) Decode(msg any) error {
	next, err := readJSONNext(r.it)
	if err != nil {
		return err
	}
	switch next {
	case jsoniter.ArrayValue:
		if r.inArray {
			return errJSONToken.New()
		}
		r.inArray = true
		if err := readJSONArray(r.it); err != nil {
			return err
		}
		fallthrough
	case jsoniter.ObjectValue:
		if err := readJSONValue(r.it, msg); err != nil {
			return err
		}
	default:
		return errJSONToken.New()
	}
	if r.inArray {
		return readJSONArray(r.it)
	}
	return nil
}

type bytesDecoder struct {
	s       *bufio.Scanner
	decoder func(string) ([]byte, error)
}

func (d *bytesDecoder) Decode(i any) error {
	buf, ok := i.(*[]byte)
	if !ok {
		panic("bytes decoder only supports *[]byte")
	}
	if !d.s.Scan() {
		return io.EOF
	}
	var err error
	*buf, err = d.decoder(strings.TrimSpace(d.s.Text()))
	if err != nil {
		return err
	}
	return nil
}

func NewBase64Decoder(r io.Reader) Decoder {
	return &bytesDecoder{
		s:       bufio.NewScanner(r),
		decoder: base64.StdEncoding.DecodeString,
	}
}

func NewHexDecoder(r io.Reader) Decoder {
	return &bytesDecoder{
		s:       bufio.NewScanner(r),
		decoder: hex.DecodeString,
	}
}
