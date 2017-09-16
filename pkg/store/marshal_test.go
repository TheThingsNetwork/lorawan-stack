// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store_test

import (
	"encoding"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/gogo/protobuf/proto"
	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

type SubSubStruct struct {
	String    string
	Int       int
	Bytes     []byte
	ByteArray byteArray
	Empty     interface{}
}
type SubStruct struct {
	String       string
	Int          int
	Bytes        []byte
	ByteArray    byteArray
	Empty        interface{}
	SubSubStruct SubSubStruct
}
type Struct struct {
	String    string
	Int       int
	Bytes     []byte
	ByteArray byteArray
	Empty     interface{}
	SubStruct SubStruct
}

type byteArray [2]byte

var (
	interfaceMapInput = map[string]interface{}{
		"string":    "string",
		"int":       42,
		"bytes":     []byte("bytes"),
		"byteArray": byteArray([2]byte{'4', '2'}),
		"sub": map[string]interface{}{
			"string":    "string",
			"int":       42,
			"bytes":     []byte("bytes"),
			"byteArray": byteArray([2]byte{'4', '2'}),
			"sub": map[string]interface{}{
				"string":    "string",
				"int":       42,
				"bytes":     []byte("bytes"),
				"byteArray": byteArray([2]byte{'4', '2'}),
			},
		},
	}

	structInput = Struct{
		String:    "string",
		Int:       42,
		Bytes:     []byte("bytes"),
		ByteArray: byteArray([2]byte{'4', '2'}),
		SubStruct: SubStruct{
			String:    "string",
			Int:       42,
			Bytes:     []byte("bytes"),
			ByteArray: byteArray([2]byte{'4', '2'}),
			SubSubStruct: SubSubStruct{
				String:    "string",
				Int:       42,
				Bytes:     []byte("bytes"),
				ByteArray: byteArray([2]byte{'4', '2'}),
			},
		},
	}
)

func TestMarshalMap(t *testing.T) {
	a := assertions.New(t)
	for _, tc := range []struct {
		input    interface{}
		expected map[string]interface{}
	}{
		{
			interfaceMapInput,
			map[string]interface{}{
				"string":            "string",
				"int":               42,
				"bytes":             []byte("bytes"),
				"byteArray":         byteArray([2]byte{'4', '2'}),
				"sub.string":        "string",
				"sub.int":           42,
				"sub.bytes":         []byte("bytes"),
				"sub.byteArray":     byteArray([2]byte{'4', '2'}),
				"sub.sub.string":    "string",
				"sub.sub.int":       42,
				"sub.sub.bytes":     []byte("bytes"),
				"sub.sub.byteArray": byteArray([2]byte{'4', '2'}),
			},
		},

		{
			structInput,
			map[string]interface{}{
				"String":                           "string",
				"Int":                              42,
				"Bytes":                            []byte("bytes"),
				"ByteArray":                        byteArray([2]byte{'4', '2'}),
				"SubStruct.String":                 "string",
				"SubStruct.Int":                    42,
				"SubStruct.Bytes":                  []byte("bytes"),
				"SubStruct.ByteArray":              byteArray([2]byte{'4', '2'}),
				"SubStruct.SubSubStruct.String":    "string",
				"SubStruct.SubSubStruct.Int":       42,
				"SubStruct.SubSubStruct.Bytes":     []byte("bytes"),
				"SubStruct.SubSubStruct.ByteArray": byteArray([2]byte{'4', '2'}),
			},
		},

		{
			struct {
				a int
				b int
			}{},
			map[string]interface{}{},
		},
		{
			struct{ time.Time }{time.Unix(42, 42)},
			map[string]interface{}{"Time": time.Unix(42, 42)},
		},
		{
			struct{ T time.Time }{time.Unix(42, 42)},
			map[string]interface{}{"T": time.Unix(42, 42)},
		},
	} {
		m := MarshalMap(tc.input)
		if !a.So(m, should.Resemble, tc.expected) {
			t.Log(pretty.Sprintf("\n%# v\n does not resemble\n %# v\n", m, tc.expected))
		}

		v := reflect.New(reflect.TypeOf(tc.input))
		a.So(UnmarshalMap(m, v.Interface()), should.BeNil)
		a.So(reflect.Indirect(v).Interface(), should.Resemble, reflect.Indirect(reflect.ValueOf(tc.input)).Interface())
	}
}

func marshalToBytes(v interface{}) (b []byte) {
	var (
		token Encoding
		err   error
	)
	switch v := v.(type) {
	case encoding.BinaryMarshaler:
		token = BinaryEncoding
		b, err = v.MarshalBinary()
	case encoding.TextMarshaler:
		token = TextEncoding
		b, err = v.MarshalText()
	case proto.Marshaler:
		token = ProtoEncoding
		b, err = v.Marshal()
	case json.Marshaler:
		token = JSONEncoding
		b, err = v.MarshalJSON()
	}
	if err != nil {
		panic(err)
	}
	return append([]byte{byte(token)}, b...)
}

func TestMarshalByteMap(t *testing.T) {
	a := assertions.New(t)
	for _, tc := range []struct {
		input    interface{}
		expected map[string][]byte
	}{
		{
			interfaceMapInput,
			map[string][]byte{
				"string":            []byte("string"),
				"int":               []byte("42"),
				"bytes":             []byte("bytes"),
				"byteArray":         []byte{'4', '2'},
				"sub.string":        []byte("string"),
				"sub.int":           []byte("42"),
				"sub.bytes":         []byte("bytes"),
				"sub.byteArray":     []byte{'4', '2'},
				"sub.sub.string":    []byte("string"),
				"sub.sub.int":       []byte("42"),
				"sub.sub.bytes":     []byte("bytes"),
				"sub.sub.byteArray": []byte{'4', '2'},
			},
		},

		{
			structInput,
			map[string][]byte{
				"String":                           []byte("string"),
				"Int":                              []byte("42"),
				"Bytes":                            []byte("bytes"),
				"ByteArray":                        []byte{'4', '2'},
				"SubStruct.String":                 []byte("string"),
				"SubStruct.Int":                    []byte("42"),
				"SubStruct.Bytes":                  []byte("bytes"),
				"SubStruct.ByteArray":              []byte{'4', '2'},
				"SubStruct.SubSubStruct.String":    []byte("string"),
				"SubStruct.SubSubStruct.Int":       []byte("42"),
				"SubStruct.SubSubStruct.Bytes":     []byte("bytes"),
				"SubStruct.SubSubStruct.ByteArray": []byte{'4', '2'},
			},
		},
		{
			struct {
				a int
				b int
			}{},
			map[string][]byte{},
		},
		{
			struct{ time.Time }{time.Unix(42, 42)},
			map[string][]byte{"Time": marshalToBytes(time.Unix(42, 42))},
		},
		{
			struct{ T time.Time }{time.Unix(42, 42)},
			map[string][]byte{"T": marshalToBytes(time.Unix(42, 42))},
		},
	} {
		m, err := MarshalByteMap(tc.input)
		a.So(err, should.BeNil)
		if !a.So(m, should.Resemble, tc.expected) {
			t.Log(pretty.Sprintf("\n%# v\n does not resemble\n %# v\n", tc.expected, m))
		}

		if reflect.DeepEqual(tc.input, interfaceMapInput) {
			// can not unmarshal into interface{} map
			continue
		}
		v := reflect.New(reflect.TypeOf(tc.input))
		a.So(UnmarshalByteMap(m, v.Interface()), should.BeNil)
		a.So(reflect.Indirect(v).Interface(), should.Resemble, reflect.Indirect(reflect.ValueOf(tc.input)).Interface())
	}
}
