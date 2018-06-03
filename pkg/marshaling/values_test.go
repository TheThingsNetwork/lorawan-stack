// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package marshaling_test

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/tinylib/msgp/msgp"
	"github.com/vmihailenco/msgpack"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/marshaling"
)

type InterfaceStructA struct {
	A int
}

type InterfaceStructB struct {
	A map[string]bool
}

type SubSubStruct struct {
	ByteArray byteArray
	Bytes     []byte
	Empty     interface{}
	Int       int
	Interface interface{}
	String    string
}
type SubStruct struct {
	ByteArray    byteArray
	Bytes        []byte
	Empty        interface{}
	Int          int
	Interface    interface{}
	String       string
	SubSubStruct SubSubStruct
}

type StructWithMap struct {
	Map map[string]string
}

type Struct struct {
	ByteArray         byteArray
	Bytes             []byte
	Empty             interface{}
	Int               int
	Int64             int64
	Uint8             uint8
	Float32           float32
	Interface         interface{}
	String            string
	StructSlice       []SubStruct
	SubStruct         SubStruct
	SubStructPtr      *SubStruct
	StructWithMap     StructWithMap
	StructWithZeroMap StructWithMap
	StructWithNilMap  StructWithMap
	NilPtr            *struct{}
}

type byteArray [2]byte

func mustToBytes(v interface{}) []byte {
	b, err := ToBytes(v)
	if err != nil {
		panic(err)
	}
	return b
}

func mustToBytesValue(v reflect.Value) []byte {
	b, err := ToBytesValue(v)
	if err != nil {
		panic(err)
	}
	return b
}

func wrapValue(v reflect.Value, t reflect.Type) reflect.Value {
	wv := reflect.New(t)
	wv.Elem().Set(v)
	return wv.Elem()
}

type ProtoMarshaler struct {
	a int
}

var _ proto.Marshaler = ProtoMarshaler{}
var _ proto.Message = &ProtoMarshaler{}

func (m *ProtoMarshaler) Reset()         { *m = ProtoMarshaler{} }
func (m *ProtoMarshaler) String() string { return proto.CompactTextString(m) }
func (*ProtoMarshaler) ProtoMessage()    {}

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

var _ MapMarshaler = CustomMarshaler{}
var _ MapUnmarshaler = &CustomMarshaler{}
var _ ByteMapMarshaler = CustomMarshaler{}
var _ ByteMapUnmarshaler = &CustomMarshaler{}

type CustomMarshaler struct {
	a uint8
	b byte
	c []byte
}

func (cm CustomMarshaler) MarshalMap() (map[string]interface{}, error) {
	return map[string]interface{}{
		"aField": cm.a,
		"bField": cm.b,
		"cField": append(cm.c, 'X'),
	}, nil
}

func (cm CustomMarshaler) MarshalByteMap() (map[string][]byte, error) {
	return map[string][]byte{
		"aField": {cm.a},
		"bField": {cm.b},
		"cField": append(cm.c, 'X', 'X'),
	}, nil
}

func (cm *CustomMarshaler) UnmarshalMap(m map[string]interface{}) error {
	*cm = CustomMarshaler{
		a: m["aField"].(uint8),
		b: m["bField"].(byte),
		c: m["cField"].([]byte),
	}
	cm.c = cm.c[:len(cm.c)-1]
	return nil
}

func (cm *CustomMarshaler) UnmarshalByteMap(m map[string][]byte) error {
	*cm = CustomMarshaler{
		a: m["aField"][0],
		b: m["bField"][0],
		c: m["cField"],
	}
	cm.c = cm.c[:len(cm.c)-2]
	return nil
}

type CustomMarshalerAB struct {
	A *CustomMarshaler
	B *CustomMarshaler
}

func (cm CustomMarshalerAB) MarshalMap() (map[string]interface{}, error) {
	return map[string]interface{}{
		"A.aField": cm.A.a,
		"A.bField": cm.A.b,
		"A.cField": append(cm.A.c, 'X'),
		"B.aField": cm.B.a,
		"B.bField": cm.B.b,
		"B.cField": append(cm.B.c, 'X'),
	}, nil
}

func (cm CustomMarshalerAB) MarshalByteMap() (map[string][]byte, error) {
	return map[string][]byte{
		"A.aField": {cm.A.a},
		"A.bField": {cm.A.b},
		"A.cField": append(cm.A.c, 'X', 'X'),
		"B.aField": {cm.B.a},
		"B.bField": {cm.B.b},
		"B.cField": append(cm.B.c, 'X', 'X'),
	}, nil
}

func (cm *CustomMarshalerAB) UnmarshalMap(m map[string]interface{}) error {
	*cm = CustomMarshalerAB{
		A: &CustomMarshaler{
			a: m["A.aField"].(uint8),
			b: m["A.bField"].(byte),
			c: m["A.cField"].([]byte),
		},
		B: &CustomMarshaler{
			a: m["B.aField"].(uint8),
			b: m["B.bField"].(byte),
			c: m["B.cField"].([]byte),
		},
	}
	cm.A.c = cm.A.c[:len(cm.A.c)-1]
	cm.B.c = cm.B.c[:len(cm.B.c)-1]
	return nil
}

func (cm *CustomMarshalerAB) UnmarshalByteMap(m map[string][]byte) error {
	*cm = CustomMarshalerAB{
		A: &CustomMarshaler{
			a: m["A.aField"][0],
			b: m["A.bField"][0],
			c: m["A.cField"],
		},
		B: &CustomMarshaler{
			a: m["B.aField"][0],
			b: m["B.bField"][0],
			c: m["B.cField"],
		},
	}
	cm.A.c = cm.A.c[:len(cm.A.c)-2]
	cm.B.c = cm.B.c[:len(cm.B.c)-2]
	return nil
}

// Trick to register types before values is declared.
var _ interface{} = func() interface{} {
	for _, v := range []interface{}{
		InterfaceStructA{},
		InterfaceStructB{},
		time.Time{},
		struct{ A int }{},
	} {
		gob.Register(v)
	}
	return nil
}()

var structValues = []struct {
	unmarshaled interface{}
	marshaled   map[string]interface{}
	bytes       map[string][]byte
}{
	{
		Struct{
			String:    "42",
			Int:       42,
			Int64:     42,
			Uint8:     42,
			Float32:   42.42,
			Interface: InterfaceStructA{42},
			Bytes:     []byte("42"),
			ByteArray: byteArray{'4', '2'},
			StructSlice: []SubStruct{
				{
					String:    "42",
					ByteArray: byteArray{'4', '2'},
				},
				{
					Int:       42,
					Interface: float64(42),
					Bytes:     []byte("42"),
				},
				{
					Interface: InterfaceStructB{map[string]bool{"42": true}},
				},
			},
			SubStruct: SubStruct{
				String:    "42",
				Int:       42,
				Interface: "42",
				Bytes:     []byte("42"),
				ByteArray: byteArray{'4', '2'},
				SubSubStruct: SubSubStruct{
					String:    "42",
					Int:       42,
					Bytes:     []byte("42"),
					ByteArray: byteArray{'4', '2'},
				},
			},
			SubStructPtr: &SubStruct{
				String:    "42",
				Int:       42,
				ByteArray: byteArray([2]byte{}),
			},
			NilPtr:            nil,
			StructWithMap:     StructWithMap{Map: map[string]string{"42": "foo"}},
			StructWithZeroMap: StructWithMap{Map: map[string]string{}},
			StructWithNilMap:  StructWithMap{Map: (map[string]string)(nil)},
		},
		map[string]interface{}{
			"ByteArray": mustToBytes(byteArray{'4', '2'}),
			"Bytes":     mustToBytes([]byte("42")),
			"Empty":     nil,
			"Float32":   float32(42.42),
			"Int":       int(42),
			"Int64":     int64(42),
			"Uint8":     uint8(42),
			"Interface": mustToBytesValue(
				wrapValue(reflect.ValueOf(InterfaceStructA{42}),
					reflect.TypeOf((*interface{})(nil)).Elem())),
			"String": string("42"),
			"StructSlice": mustToBytes([]SubStruct{
				{
					String:    "42",
					ByteArray: byteArray{'4', '2'},
				},
				{
					Int:       42,
					Interface: float64(42),
					Bytes:     []byte("42"),
				},
				{
					Interface: InterfaceStructB{map[string]bool{"42": true}},
				},
			}),
			"NilPtr":                nil,
			"StructWithMap.Map":     mustToBytes(map[string]string{"42": "foo"}),
			"StructWithZeroMap.Map": mustToBytes(map[string]string{}),
			"StructWithNilMap.Map":  nil,
			"SubStruct.ByteArray":   mustToBytes(byteArray{'4', '2'}),
			"SubStruct.Bytes":       mustToBytes([]byte("42")),
			"SubStruct.Empty":       nil,
			"SubStruct.Int":         int(42),
			"SubStruct.Interface": mustToBytesValue(
				wrapValue(reflect.ValueOf("42"),
					reflect.TypeOf((*interface{})(nil)).Elem())),
			"SubStruct.String":                    string("42"),
			"SubStruct.SubSubStruct.ByteArray":    mustToBytes(byteArray{'4', '2'}),
			"SubStruct.SubSubStruct.Bytes":        mustToBytes([]byte("42")),
			"SubStruct.SubSubStruct.Empty":        nil,
			"SubStruct.SubSubStruct.Int":          int(42),
			"SubStruct.SubSubStruct.Interface":    nil,
			"SubStruct.SubSubStruct.String":       string("42"),
			"SubStructPtr.ByteArray":              mustToBytes(byteArray{}),
			"SubStructPtr.Bytes":                  nil,
			"SubStructPtr.Empty":                  nil,
			"SubStructPtr.Int":                    int(42),
			"SubStructPtr.Interface":              nil,
			"SubStructPtr.String":                 string("42"),
			"SubStructPtr.SubSubStruct.ByteArray": mustToBytes(byteArray{}),
			"SubStructPtr.SubSubStruct.Bytes":     nil,
			"SubStructPtr.SubSubStruct.Empty":     nil,
			"SubStructPtr.SubSubStruct.Int":       int(0),
			"SubStructPtr.SubSubStruct.Interface": nil,
			"SubStructPtr.SubSubStruct.String":    string(""),
		},

		map[string][]byte{
			"ByteArray": mustToBytes(byteArray{'4', '2'}),
			"Bytes":     mustToBytes([]byte("42")),
			"Empty":     nil,
			"Interface": mustToBytesValue(
				wrapValue(reflect.ValueOf(InterfaceStructA{42}),
					reflect.TypeOf((*interface{})(nil)).Elem())),
			"Int":                   mustToBytes(int(42)),
			"Float32":               mustToBytes(float32(42.42)),
			"Int64":                 mustToBytes(int64(42)),
			"Uint8":                 mustToBytes(uint8(42)),
			"String":                mustToBytes("42"),
			"NilPtr":                nil,
			"StructWithMap.Map":     mustToBytes(map[string]string{"42": "foo"}),
			"StructWithZeroMap.Map": mustToBytes(map[string]string{}),
			"StructWithNilMap.Map":  nil,
			"StructSlice": mustToBytes([]SubStruct{
				{
					String:    "42",
					ByteArray: byteArray{'4', '2'},
				},
				{
					Int:       42,
					Interface: float64(42),
					Bytes:     []byte("42"),
				},
				{
					Interface: InterfaceStructB{map[string]bool{"42": true}},
				},
			}),
			"SubStruct.ByteArray": mustToBytes(byteArray{'4', '2'}),
			"SubStruct.Bytes":     mustToBytes([]byte("42")),
			"SubStruct.Int":       mustToBytes(int(42)),
			"SubStruct.Interface": mustToBytesValue(
				wrapValue(reflect.ValueOf("42"),
					reflect.TypeOf((*interface{})(nil)).Elem())),
			"SubStruct.Empty":                     nil,
			"SubStruct.String":                    mustToBytes("42"),
			"SubStruct.SubSubStruct.ByteArray":    mustToBytes(byteArray{'4', '2'}),
			"SubStruct.SubSubStruct.Bytes":        mustToBytes([]byte("42")),
			"SubStruct.SubSubStruct.Empty":        nil,
			"SubStruct.SubSubStruct.Int":          mustToBytes(int(42)),
			"SubStruct.SubSubStruct.Interface":    nil,
			"SubStruct.SubSubStruct.String":       mustToBytes("42"),
			"SubStructPtr.ByteArray":              mustToBytes(byteArray{}),
			"SubStructPtr.Bytes":                  nil,
			"SubStructPtr.Empty":                  nil,
			"SubStructPtr.Int":                    mustToBytes(int(42)),
			"SubStructPtr.Interface":              nil,
			"SubStructPtr.String":                 mustToBytes("42"),
			"SubStructPtr.SubSubStruct.ByteArray": mustToBytes(byteArray{}),
			"SubStructPtr.SubSubStruct.Bytes":     nil,
			"SubStructPtr.SubSubStruct.Empty":     nil,
			"SubStructPtr.SubSubStruct.Int":       mustToBytes(0),
			"SubStructPtr.SubSubStruct.Interface": nil,
			"SubStructPtr.SubSubStruct.String":    mustToBytes(""),
		},
	},
	{
		struct {
			EmptyArray [2]uint
			EmptyBytes []byte
			EmptyInts  []int
			NilBytes   []byte
		}{
			EmptyBytes: make([]byte, 0),
			EmptyInts:  make([]int, 0),
			NilBytes:   ([]byte)(nil),
		},
		map[string]interface{}{
			"EmptyArray": mustToBytes([2]uint{}),
			"EmptyBytes": mustToBytes(make([]byte, 0)),
			"EmptyInts":  mustToBytes(make([]int, 0)),
			"NilBytes":   nil,
		},
		map[string][]byte{
			"EmptyArray": mustToBytes([2]uint{}),
			"EmptyBytes": mustToBytes(make([]byte, 0)),
			"EmptyInts":  mustToBytes(make([]int, 0)),
			"NilBytes":   nil,
		},
	},
	{
		struct {
			a int
			b int
		}{},
		(map[string]interface{})(nil),
		(map[string][]byte)(nil),
	},
	{
		struct {
			A interface{}
			B interface{}
		}{
			42,
			struct{ A int }{42},
		},
		map[string]interface{}{
			"A": mustToBytesValue(
				wrapValue(reflect.ValueOf(42),
					reflect.TypeOf((*interface{})(nil)).Elem())),
			"B": mustToBytesValue(
				wrapValue(reflect.ValueOf(struct{ A int }{42}),
					reflect.TypeOf((*interface{})(nil)).Elem())),
		},
		map[string][]byte{
			"A": mustToBytesValue(
				wrapValue(reflect.ValueOf(42),
					reflect.TypeOf((*interface{})(nil)).Elem())),
			"B": mustToBytesValue(
				wrapValue(reflect.ValueOf(struct{ A int }{42}),
					reflect.TypeOf((*interface{})(nil)).Elem())),
		},
	},
	{
		struct{ time.Time }{time.Unix(42, 42).UTC()},
		map[string]interface{}{"Time": mustToBytes(time.Unix(42, 42).UTC())},
		map[string][]byte{"Time": mustToBytes(time.Unix(42, 42).UTC())},
	},
	{
		struct{ T time.Time }{time.Unix(42, 42).UTC()},
		map[string]interface{}{"T": mustToBytes(time.Unix(42, 42).UTC())},
		map[string][]byte{"T": mustToBytes(time.Unix(42, 42).UTC())},
	},
	{
		struct{ Interfaces []interface{} }{[]interface{}{
			nil,
			nil,
			nil,
			time.Time{},
			&time.Time{},
			struct{ A int }{42},
		}},
		map[string]interface{}{
			"Interfaces": mustToBytes([]interface{}{
				nil,
				nil,
				nil,
				time.Time{},
				&time.Time{},
				struct{ A int }{42},
			}),
		},
		map[string][]byte{
			"Interfaces": mustToBytes([]interface{}{
				nil,
				nil,
				nil,
				time.Time{},
				&time.Time{},
				struct{ A int }{42},
			}),
		},
	},
	{
		struct {
			A *ProtoMarshaler
		}{
			&ProtoMarshaler{42},
		},
		map[string]interface{}{
			"A": mustToBytes(ProtoMarshaler{42}),
		},
		map[string][]byte{
			"A": mustToBytes(ProtoMarshaler{42}),
		},
	},
	{
		CustomMarshaler{
			a: 42,
			b: 43,
			c: []byte("foo"),
		},
		map[string]interface{}{
			"aField": uint8(42),
			"bField": byte(43),
			"cField": []byte("fooX"),
		},
		map[string][]byte{
			"aField": {42},
			"bField": {43},
			"cField": []byte("fooXX"),
		},
	},
	{
		CustomMarshalerAB{
			&CustomMarshaler{
				a: 42,
				b: 43,
				c: []byte("foo"),
			},
			&CustomMarshaler{
				a: 4,
				b: 5,
				c: []byte("bar"),
			},
		},
		map[string]interface{}{
			"A.aField": uint8(42),
			"A.bField": byte(43),
			"A.cField": []byte("fooX"),
			"B.aField": uint8(4),
			"B.bField": byte(5),
			"B.cField": []byte("barX"),
		},
		map[string][]byte{
			"A.aField": {42},
			"A.bField": {43},
			"A.cField": []byte("fooXX"),
			"B.aField": {4},
			"B.bField": {5},
			"B.cField": []byte("barXX"),
		},
	},
}

func gobEncoded(v interface{}) []byte {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(v); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func msgPackEncoded(v interface{}) (b []byte) {
	var err error
	if m, ok := v.(msgp.Marshaler); ok {
		b, err = m.MarshalMsg(nil)
	} else {
		b, err = msgpack.Marshal(v)
	}

	if err != nil {
		panic(err)
	}
	return b
}

var byteValues = []struct {
	value interface{}
	bytes []byte
}{
	{
		int(42),
		append([]byte{byte(GobEncoding)}, gobEncoded(int(42))...),
	},
	{
		int8(42),
		append([]byte{byte(GobEncoding)}, gobEncoded(int8(42))...),
	},
	{
		int16(42),
		append([]byte{byte(GobEncoding)}, gobEncoded(int16(42))...),
	},
	{
		int32(42),
		append([]byte{byte(GobEncoding)}, gobEncoded(int32(42))...),
	},
	{
		int64(42),
		append([]byte{byte(GobEncoding)}, gobEncoded(int64(42))...),
	},
	{
		uint(42),
		append([]byte{byte(GobEncoding)}, gobEncoded(uint(42))...),
	},
	{
		uint8(42),
		append([]byte{byte(GobEncoding)}, gobEncoded(uint8(42))...),
	},
	{
		uint16(42),
		append([]byte{byte(GobEncoding)}, gobEncoded(uint16(42))...),
	},
	{
		uint32(42),
		append([]byte{byte(GobEncoding)}, gobEncoded(uint32(42))...),
	},
	{
		uint64(42),
		append([]byte{byte(GobEncoding)}, gobEncoded(uint64(42))...),
	},
	{
		float32(42),
		append([]byte{byte(GobEncoding)}, gobEncoded(float32(42))...),
	},
	{
		float64(42),
		append([]byte{byte(GobEncoding)}, gobEncoded(float64(42))...),
	},
	{
		[]byte("42"),
		append([]byte{byte(GobEncoding)}, gobEncoded([]byte("42"))...),
	},
	{
		make([]byte, 0),
		append([]byte{byte(GobEncoding)}, gobEncoded(make([]byte, 0))...),
	},
	{
		make([]int, 0),
		append([]byte{byte(GobEncoding)}, gobEncoded(make([]int, 0))...),
	},
	{
		make(map[string]interface{}),
		append([]byte{byte(GobEncoding)}, gobEncoded(make(map[string]interface{}))...),
	},
	{
		map[string]interface{}{"1": 42, "2": uint8(32), "3": int64(44), "hey": "foo", "bar": []byte("baz")},
		append([]byte{byte(GobEncoding)}, gobEncoded(map[string]interface{}{"1": 42, "2": uint8(32), "3": int64(44), "hey": "foo", "bar": []byte("baz")})...),
	},
	{
		"42",
		append([]byte{byte(GobEncoding)}, gobEncoded("42")...),
	},
	{
		[]interface{}{1, 2},
		append([]byte{byte(GobEncoding)}, gobEncoded([]interface{}{1, 2})...),
	},
	{
		&map[string]interface{}{"asd": uint32(42)},
		append([]byte{byte(GobEncoding)}, gobEncoded(&map[string]interface{}{"asd": uint32(42)})...),
	},
	{
		struct{ A int }{42},
		append([]byte{byte(GobEncoding)}, gobEncoded(struct{ A int }{42})...),
	},
	{
		&struct{ A int }{42},
		append([]byte{byte(GobEncoding)}, gobEncoded(&struct{ A int }{42})...),
	},
	{
		struct{ V interface{} }{struct{ A int }{42}},
		append([]byte{byte(GobEncoding)}, gobEncoded(struct{ V interface{} }{struct{ A int }{42}})...),
	},
	{
		nil,
		nil,
	},
}
