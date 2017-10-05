// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"strconv"
	strings "strings"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/gogo/protobuf/jsonpb"
)

// ParseRight parses the string specified into a Right.
func ParseRight(str string) (Right, error) {
	val, ok := Right_value["RIGHT_"+strings.ToUpper(strings.Replace(str, ":", "_", -1))]
	if !ok {
		val, ok = Right_value[str]
		if !ok {
			return -1, errors.Errorf("Could not parse right `%s`", str)
		}
	}
	return Right(val), nil
}

// TextString returns a textual string representation of the right.
func (r Right) TextString() string {
	str, ok := Right_name[int32(r)]
	if ok {
		return strings.ToLower(strings.Replace(strings.TrimPrefix(str, "RIGHT_"), "_", ":", -1))
	}
	return strconv.Itoa(int(r))
}

// MarshalText implements encoding.TextMarshaler interface.
func (r Right) MarshalText() ([]byte, error) {
	return []byte(r.TextString()), nil
}

// MarshalJSON implements json.Marshaler interface.
func (r Right) MarshalJSON() ([]byte, error) {
	txt, err := r.MarshalText()
	if err != nil {
		return nil, err
	}
	return []byte("\"" + string(txt) + "\""), nil
}

// MarshalJSONPB implements jsonpb.JSONPBMarshaler interface.
func (r Right) MarshalJSONPB(m *jsonpb.Marshaler) ([]byte, error) {
	if m.EnumsAsInts {
		return []byte("\"" + strconv.Itoa(int(r)) + "\""), nil
	}
	return r.MarshalJSON()
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (r *Right) UnmarshalText(b []byte) (err error) {
	*r, err = ParseRight(string(b))
	return
}

// MarshalJSON implements json.Unmarshaler interface.
func (r *Right) UnmarshalJSON(b []byte) error {
	return r.UnmarshalText(b[1 : len(b)-1])
}

// UnmarshalJSONPB implements jsonpb.JSONPBUnmarshaler interface.
func (r *Right) UnmarshalJSONPB(m *jsonpb.Unmarshaler, b []byte) error {
	return r.UnmarshalJSON(b)
}
