// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store_test

import (
	"strconv"
	"testing"

	. "github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestMarshalMap(t *testing.T) {
	for i, v := range values {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			m := MarshalMap(v.unmarshaled)
			if !a.So(m, should.Resemble, v.marshaled) {
				pretty.Ldiff(t, m, v.marshaled)
			}
		})
	}
}

func TestMarshalByteMap(t *testing.T) {
	for i, v := range values {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			m, err := MarshalByteMap(v.unmarshaled)
			a.So(err, should.BeNil)
			if !a.So(m, should.Resemble, v.bytes) {
				pretty.Ldiff(t, m, v.bytes)
			}
		})
	}
}
