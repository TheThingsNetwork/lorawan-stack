// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store_test

import (
	"testing"

	. "github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestUnmarshal(t *testing.T) {
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

	input := map[string]interface{}{
		"String":                        "string",
		"Int":                           42,
		"SubStruct.String":              "string",
		"SubStruct.Int":                 42,
		"SubStruct.SubSubStruct.String": "string",
		"SubStruct.SubSubStruct.Int":    42,
	}
	expected := Struct{
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
	}

	var v Struct
	err := UnmarshalMap(input, &v)
	a.So(err, should.BeNil)
	if !a.So(v, should.Resemble, expected) {
		t.Log(pretty.Sprintf("\n%# v\n does not resemble\n %# v\n", v, expected))
	}
}
