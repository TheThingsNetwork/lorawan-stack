// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store_test

import (
	"encoding/json"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	. "github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/gogo/protobuf/proto"
	"github.com/mitchellh/mapstructure"
)

type SubSubStruct struct {
	ByteArray byteArray
	Bytes     []byte
	Empty     interface{}
	Int       int
	String    string
}
type SubStruct struct {
	ByteArray    byteArray
	Bytes        []byte
	Empty        interface{}
	Int          int
	String       string
	SubSubStruct SubSubStruct
}
type Struct struct {
	ByteArray   byteArray
	Bytes       []byte
	Empty       interface{}
	Int         int
	String      string
	StructSlice []SubStruct
	SubStruct   SubStruct
}

type byteArray [2]byte

func mustToBytes(v interface{}) []byte {
	b, err := ToBytes(v)
	if err != nil {
		panic(err)
	}
	return b
}

type ProtoMarshaler struct {
	a int
}

var _ proto.Marshaler = ProtoMarshaler{}
var _ proto.Unmarshaler = &ProtoMarshaler{}

func (m ProtoMarshaler) Marshal() ([]byte, error) {
	return []byte{byte(m.a), byte(ProtoEncoding)}, nil
}

func (m *ProtoMarshaler) Unmarshal(b []byte) error {
	if len(b) != 2 {
		return errors.Errorf("Encoded length must be 2, got %d", len(b))
	}
	if Encoding(b[1]) != ProtoEncoding {
		return errors.Errorf("Second byte must be %d, got %d", ProtoEncoding, b[1])
	}
	*m = ProtoMarshaler{
		a: int(b[0]),
	}
	return nil
}

type JSONMarshaler struct {
	a int
}

var _ json.Marshaler = JSONMarshaler{}
var _ json.Unmarshaler = &JSONMarshaler{}

func (m JSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte{byte(m.a), byte(JSONEncoding)}, nil
}

func (m *JSONMarshaler) UnmarshalJSON(b []byte) error {
	if len(b) != 2 {
		return errors.Errorf("Encoded length must be 2, got %d", len(b))
	}
	if Encoding(b[1]) != JSONEncoding {
		return errors.Errorf("Second byte must be %d, got %d", JSONEncoding, b[1])
	}
	*m = JSONMarshaler{
		a: int(b[0]),
	}
	return nil
}

type ProtoJSONMarshaler struct {
	a int
}

var _ proto.Marshaler = ProtoJSONMarshaler{}
var _ proto.Unmarshaler = &ProtoJSONMarshaler{}
var _ json.Marshaler = ProtoJSONMarshaler{}
var _ json.Unmarshaler = &ProtoJSONMarshaler{}

func (m ProtoJSONMarshaler) Marshal() ([]byte, error) {
	return []byte{byte(m.a), byte(ProtoEncoding)}, nil
}

func (m *ProtoJSONMarshaler) Unmarshal(b []byte) error {
	if len(b) != 2 {
		return errors.Errorf("Encoded length must be 2, got %d", len(b))
	}
	if Encoding(b[1]) != ProtoEncoding {
		return errors.Errorf("Second byte must be %d, got %d", ProtoEncoding, b[1])
	}
	*m = ProtoJSONMarshaler{
		a: int(b[0]),
	}
	return nil
}

func (m ProtoJSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte{byte(m.a), byte(JSONEncoding)}, nil
}

func (m *ProtoJSONMarshaler) UnmarshalJSON(b []byte) error {
	if len(b) != 2 {
		return errors.Errorf("Encoded length must be 2, got %d", len(b))
	}
	if Encoding(b[1]) != JSONEncoding {
		return errors.Errorf("Second byte must be %d, got %d", JSONEncoding, b[1])
	}
	*m = ProtoJSONMarshaler{
		a: int(b[0]),
	}
	return nil
}

var values = []struct {
	unmarshaled interface{}
	marshaled   map[string]interface{}
	bytes       map[string][]byte
	decodeHooks []mapstructure.DecodeHookFunc
}{
	{
		Struct{
			String:    "42",
			Int:       42,
			Bytes:     []byte("42"),
			ByteArray: byteArray([2]byte{'4', '2'}),
			StructSlice: []SubStruct{
				{
					String:    "42",
					ByteArray: byteArray([2]byte{'4', '2'}),
				},
				{
					Int:   42,
					Bytes: []byte("42"),
				},
			},
			SubStruct: SubStruct{
				String:    "42",
				Int:       42,
				Bytes:     []byte("42"),
				ByteArray: byteArray([2]byte{'4', '2'}),
				SubSubStruct: SubSubStruct{
					String:    "42",
					Int:       42,
					Bytes:     []byte("42"),
					ByteArray: byteArray([2]byte{'4', '2'}),
				},
			},
		},
		map[string]interface{}{
			"ByteArray":                        byteArray([2]byte{'4', '2'}),
			"Bytes":                            []byte("42"),
			"Int":                              int(42),
			"String":                           string("42"),
			"StructSlice.0.ByteArray":          byteArray([2]byte{'4', '2'}),
			"StructSlice.0.String":             string("42"),
			"StructSlice.1.ByteArray":          byteArray([2]byte{0, 0}),
			"StructSlice.1.Bytes":              []byte("42"),
			"StructSlice.1.Int":                int(42),
			"SubStruct.ByteArray":              byteArray([2]byte{'4', '2'}),
			"SubStruct.Bytes":                  []byte("42"),
			"SubStruct.Int":                    int(42),
			"SubStruct.String":                 string("42"),
			"SubStruct.SubSubStruct.ByteArray": byteArray([2]byte{'4', '2'}),
			"SubStruct.SubSubStruct.Bytes":     []byte("42"),
			"SubStruct.SubSubStruct.Int":       int(42),
			"SubStruct.SubSubStruct.String":    string("42"),
		},

		map[string][]byte{
			"ByteArray":                        mustToBytes("42"),
			"Bytes":                            mustToBytes("42"),
			"Int":                              mustToBytes(42),
			"String":                           mustToBytes("42"),
			"StructSlice.0.ByteArray":          mustToBytes("42"),
			"StructSlice.0.String":             mustToBytes("42"),
			"StructSlice.1.ByteArray":          mustToBytes([2]byte{0, 0}),
			"StructSlice.1.Bytes":              mustToBytes("42"),
			"StructSlice.1.Int":                mustToBytes(42),
			"SubStruct.ByteArray":              mustToBytes("42"),
			"SubStruct.Bytes":                  mustToBytes("42"),
			"SubStruct.Int":                    mustToBytes(42),
			"SubStruct.String":                 mustToBytes("42"),
			"SubStruct.SubSubStruct.ByteArray": mustToBytes("42"),
			"SubStruct.SubSubStruct.Bytes":     mustToBytes("42"),
			"SubStruct.SubSubStruct.Int":       mustToBytes(42),
			"SubStruct.SubSubStruct.String":    mustToBytes("42"),
		},
		nil,
	},
	{
		struct {
			a int
			b int
		}{},
		map[string]interface{}{},
		map[string][]byte{},
		nil,
	},
	{
		struct{ time.Time }{time.Unix(42, 42).UTC()},
		map[string]interface{}{"Time": mustToBytes(time.Unix(42, 42).UTC())},
		map[string][]byte{"Time": mustToBytes(time.Unix(42, 42).UTC())},
		nil,
	},
	{
		struct{ T time.Time }{time.Unix(42, 42).UTC()},
		map[string]interface{}{"T": mustToBytes(time.Unix(42, 42).UTC())},
		map[string][]byte{"T": mustToBytes(time.Unix(42, 42).UTC())},
		nil,
	},
	{
		struct{ Interfaces []interface{} }{[]interface{}{
			nil,
			(*time.Time)(nil),
			(*struct{})(nil),
			time.Time{},
			&time.Time{},
			struct{ A int }{42},
		}},
		map[string]interface{}{
			"Interfaces.0": nil,
			"Interfaces.1": nil,
			"Interfaces.2": nil,
			"Interfaces.3": nil,
			"Interfaces.4": nil,
			"Interfaces.5": mustToBytes(struct{ A int }{42}),
		},
		map[string][]byte{
			"Interfaces.0": mustToBytes(nil),
			"Interfaces.1": mustToBytes(nil),
			"Interfaces.2": mustToBytes(nil),
			"Interfaces.3": mustToBytes(nil),
			"Interfaces.4": mustToBytes(nil),
			"Interfaces.5": mustToBytes(struct{ A int }{42}),
		},
		nil,
	},
	{
		struct {
			A *ProtoMarshaler
			B *JSONMarshaler
			C *ProtoJSONMarshaler
		}{
			&ProtoMarshaler{42},
			&JSONMarshaler{42},
			&ProtoJSONMarshaler{42},
		},
		map[string]interface{}{
			"A": mustToBytes(ProtoMarshaler{42}),
			"B": mustToBytes(JSONMarshaler{42}),
			"C": mustToBytes(ProtoJSONMarshaler{42}),
		},
		map[string][]byte{
			"A": mustToBytes(ProtoMarshaler{42}),
			"B": mustToBytes(JSONMarshaler{42}),
			"C": mustToBytes(ProtoJSONMarshaler{42}),
		},
		nil,
	},
}
