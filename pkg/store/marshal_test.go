// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store_test

import (
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestMarshalMap(t *testing.T) {
	type SubSubStruct struct {
		String string
		Int    int
	}
	type SubStruct struct {
		String       string
		Int          int
		SubSubStruct SubSubStruct
	}
	type Struct struct {
		String    string
		Int       int
		SubStruct SubStruct
	}

	a := assertions.New(t)
	for _, tc := range []struct {
		input    interface{}
		expected map[string]interface{}
	}{
		{
			map[string]interface{}{
				"string": "string",
				"int":    42,
				"sub": map[string]interface{}{
					"string": "string",
					"int":    42,
					"sub": map[string]interface{}{
						"string": "string",
						"int":    42,
					},
				},
			},
			map[string]interface{}{
				"string":         "string",
				"int":            42,
				"sub.string":     "string",
				"sub.int":        42,
				"sub.sub.string": "string",
				"sub.sub.int":    42,
			},
		},

		{
			Struct{
				"string",
				42,
				SubStruct{
					"string",
					42,
					SubSubStruct{
						"string",
						42,
					},
				},
			},
			map[string]interface{}{
				"String":                        "string",
				"Int":                           42,
				"SubStruct.String":              "string",
				"SubStruct.Int":                 42,
				"SubStruct.SubSubStruct.String": "string",
				"SubStruct.SubSubStruct.Int":    42,
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
		var ret map[string]interface{}
		if a.So(func() { ret = Marshal(tc.input) }, should.NotPanic) {
			if !a.So(ret, should.Resemble, tc.expected) {
				t.Log(pretty.Sprintf("\n%# v\n does not resemble\n %# v\n", tc.expected, ret))
			}
		}
	}
}
