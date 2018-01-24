// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store_test

import (
	"time"

	. "github.com/TheThingsNetwork/ttn/pkg/store"
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

var values = []struct {
	unmarshaled interface{}
	marshaled   map[string]interface{}
	bytes       map[string][]byte
}{
	{
		map[string]interface{}{
			"string":    "42",
			"int":       42,
			"bytes":     []byte("42"),
			"byteArray": byteArray([2]byte{'4', '2'}),
			"2dSlice":   [][]int{{4, 2}, {42}},
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
		},
		map[string]interface{}{
			"2dSlice.0.0":         int(4),
			"2dSlice.0.1":         int(2),
			"2dSlice.1.0":         int(42),
			"byteArray.0":         byte(4 + '0'),
			"byteArray.1":         byte(2 + '0'),
			"bytes.0":             byte(4 + '0'),
			"bytes.1":             byte(2 + '0'),
			"int":                 int(42),
			"string":              string("42"),
			"sub.byteArray.0":     byte(4 + '0'),
			"sub.byteArray.1":     byte(2 + '0'),
			"sub.bytes.0":         byte(4 + '0'),
			"sub.bytes.1":         byte(2 + '0'),
			"sub.int":             int(42),
			"sub.string":          string("42"),
			"sub.sub.byteArray.0": byte(4 + '0'),
			"sub.sub.byteArray.1": byte(2 + '0'),
			"sub.sub.bytes.0":     byte(4 + '0'),
			"sub.sub.bytes.1":     byte(2 + '0'),
			"sub.sub.int":         int(42),
			"sub.sub.string":      string("42"),
		},
		map[string][]byte{
			"2dSlice.0.0":       mustToBytes(4),
			"2dSlice.0.1":       mustToBytes(2),
			"2dSlice.1.0":       mustToBytes(42),
			"byteArray":         mustToBytes("42"),
			"bytes":             mustToBytes("42"),
			"int":               mustToBytes(42),
			"string":            mustToBytes("42"),
			"sub.byteArray":     mustToBytes("42"),
			"sub.bytes":         mustToBytes("42"),
			"sub.int":           mustToBytes(42),
			"sub.string":        mustToBytes("42"),
			"sub.sub.byteArray": mustToBytes("42"),
			"sub.sub.bytes":     mustToBytes("42"),
			"sub.sub.int":       mustToBytes(42),
			"sub.sub.string":    mustToBytes("42"),
		},
	},
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
			"ByteArray.0":                        byte(4 + '0'),
			"ByteArray.1":                        byte(2 + '0'),
			"Bytes.0":                            byte(4 + '0'),
			"Bytes.1":                            byte(2 + '0'),
			"Int":                                int(42),
			"String":                             string("42"),
			"StructSlice.0.ByteArray.0":          byte(4 + '0'),
			"StructSlice.0.ByteArray.1":          byte(2 + '0'),
			"StructSlice.0.String":               string("42"),
			"StructSlice.1.Bytes.0":              byte(4 + '0'),
			"StructSlice.1.Bytes.1":              byte(2 + '0'),
			"StructSlice.1.Int":                  int(42),
			"SubStruct.ByteArray.0":              byte(4 + '0'),
			"SubStruct.ByteArray.1":              byte(2 + '0'),
			"SubStruct.Bytes.0":                  byte(4 + '0'),
			"SubStruct.Bytes.1":                  byte(2 + '0'),
			"SubStruct.Int":                      int(42),
			"SubStruct.String":                   string("42"),
			"SubStruct.SubSubStruct.ByteArray.0": byte(4 + '0'),
			"SubStruct.SubSubStruct.ByteArray.1": byte(2 + '0'),
			"SubStruct.SubSubStruct.Bytes.0":     byte(4 + '0'),
			"SubStruct.SubSubStruct.Bytes.1":     byte(2 + '0'),
			"SubStruct.SubSubStruct.Int":         int(42),
			"SubStruct.SubSubStruct.String":      string("42"),
		},

		map[string][]byte{
			"ByteArray":                        mustToBytes("42"),
			"Bytes":                            mustToBytes("42"),
			"Int":                              mustToBytes(42),
			"String":                           mustToBytes("42"),
			"StructSlice.0.ByteArray":          mustToBytes("42"),
			"StructSlice.0.String":             mustToBytes("42"),
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
	},
	{
		struct {
			a int
			b int
		}{},
		map[string]interface{}{},
		map[string][]byte{},
	},
	{
		struct{ time.Time }{time.Unix(42, 42).UTC()},
		map[string]interface{}{"Time": time.Unix(42, 42).UTC()},
		map[string][]byte{"Time": mustToBytes(time.Unix(42, 42).UTC())},
	},
	{
		struct{ T time.Time }{time.Unix(42, 42).UTC()},
		map[string]interface{}{"T": time.Unix(42, 42).UTC()},
		map[string][]byte{"T": mustToBytes(time.Unix(42, 42).UTC())},
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
			"Interfaces.0":   nil,
			"Interfaces.1":   nil,
			"Interfaces.2":   nil,
			"Interfaces.3":   time.Time{},
			"Interfaces.4":   time.Time{},
			"Interfaces.5.A": 42,
		},
		map[string][]byte{
			"Interfaces.0":   mustToBytes(nil),
			"Interfaces.1":   mustToBytes(nil),
			"Interfaces.2":   mustToBytes(nil),
			"Interfaces.3":   mustToBytes(time.Time{}),
			"Interfaces.4":   mustToBytes(time.Time{}),
			"Interfaces.5.A": mustToBytes(42),
		},
	},
}
