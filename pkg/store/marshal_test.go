// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store_test

import (
	"reflect"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/pkg/store"
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
		"string":    "42",
		"int":       42,
		"bytes":     []byte("42"),
		"byteArray": byteArray([2]byte{'4', '2'}),
		"sub": map[string]interface{}{
			"string":    "42",
			"int":       42,
			"bytes":     []byte("42"),
			"byteArray": byteArray([2]byte{'4', '2'}),
			"sub": map[string]interface{}{
				"string":    "42",
				"int":       42,
				"bytes":     []byte("42"),
				"byteArray": byteArray([2]byte{'4', '2'}),
			},
		},
	}

	structInput = Struct{
		String:    "42",
		Int:       42,
		Bytes:     []byte("42"),
		ByteArray: byteArray([2]byte{'4', '2'}),
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
				"string":            interfaceMapInput["string"],
				"int":               interfaceMapInput["int"],
				"bytes":             interfaceMapInput["bytes"],
				"byteArray":         interfaceMapInput["byteArray"],
				"sub.string":        interfaceMapInput["sub"].(map[string]interface{})["string"],
				"sub.int":           interfaceMapInput["sub"].(map[string]interface{})["int"],
				"sub.bytes":         interfaceMapInput["sub"].(map[string]interface{})["bytes"],
				"sub.byteArray":     interfaceMapInput["sub"].(map[string]interface{})["byteArray"],
				"sub.sub.string":    interfaceMapInput["sub"].(map[string]interface{})["sub"].(map[string]interface{})["string"],
				"sub.sub.int":       interfaceMapInput["sub"].(map[string]interface{})["sub"].(map[string]interface{})["int"],
				"sub.sub.bytes":     interfaceMapInput["sub"].(map[string]interface{})["sub"].(map[string]interface{})["bytes"],
				"sub.sub.byteArray": interfaceMapInput["sub"].(map[string]interface{})["sub"].(map[string]interface{})["byteArray"],
			},
		},

		{
			structInput,
			map[string]interface{}{
				"String":                           structInput.String,
				"Int":                              structInput.Int,
				"Bytes":                            structInput.Bytes,
				"ByteArray":                        structInput.ByteArray,
				"SubStruct.String":                 structInput.SubStruct.String,
				"SubStruct.Int":                    structInput.SubStruct.Int,
				"SubStruct.Bytes":                  structInput.SubStruct.Bytes,
				"SubStruct.ByteArray":              structInput.SubStruct.ByteArray,
				"SubStruct.SubSubStruct.String":    structInput.SubStruct.SubSubStruct.String,
				"SubStruct.SubSubStruct.Int":       structInput.SubStruct.SubSubStruct.Int,
				"SubStruct.SubSubStruct.Bytes":     structInput.SubStruct.SubSubStruct.Bytes,
				"SubStruct.SubSubStruct.ByteArray": structInput.SubStruct.SubSubStruct.ByteArray,
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

func mustToBytes(v interface{}) []byte {
	b, err := ToBytes(v)
	if err != nil {
		panic(err)
	}
	return b
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
				"string":            mustToBytes("42"),
				"int":               mustToBytes(42),
				"bytes":             mustToBytes("42"),
				"byteArray":         mustToBytes("42"),
				"sub.string":        mustToBytes("42"),
				"sub.int":           mustToBytes(42),
				"sub.bytes":         mustToBytes("42"),
				"sub.byteArray":     mustToBytes("42"),
				"sub.sub.string":    mustToBytes("42"),
				"sub.sub.int":       mustToBytes(42),
				"sub.sub.bytes":     mustToBytes("42"),
				"sub.sub.byteArray": mustToBytes("42"),
			},
		},

		{
			structInput,
			map[string][]byte{
				"String":                           mustToBytes("42"),
				"Int":                              mustToBytes(42),
				"Bytes":                            mustToBytes("42"),
				"ByteArray":                        mustToBytes("42"),
				"SubStruct.String":                 mustToBytes("42"),
				"SubStruct.Int":                    mustToBytes(42),
				"SubStruct.Bytes":                  mustToBytes("42"),
				"SubStruct.ByteArray":              mustToBytes("42"),
				"SubStruct.SubSubStruct.String":    mustToBytes("42"),
				"SubStruct.SubSubStruct.Int":       mustToBytes(42),
				"SubStruct.SubSubStruct.Bytes":     mustToBytes("42"),
				"SubStruct.SubSubStruct.ByteArray": mustToBytes("42"),
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
			map[string][]byte{"Time": mustToBytes(time.Unix(42, 42))},
		},
		{
			struct{ T time.Time }{time.Unix(42, 42)},
			map[string][]byte{"T": mustToBytes(time.Unix(42, 42))},
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
