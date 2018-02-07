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

			m, err := MarshalMap(v.unmarshaled)
			a.So(err, should.BeNil)
			if !a.So(m, should.Resemble, v.marshaled) {
				pretty.Ldiff(t, m, v.marshaled)
			}
		})
	}
	t.Run("interface", func(t *testing.T) {
		a := assertions.New(t)

		v := struct {
			A interface{}
		}{
			map[string]interface{}{
				"A": "foo",
				"B": 42,
			},
		}
		m, err := MarshalMap(v)
		a.So(err, should.BeNil)
		a.So(m, should.Resemble, map[string]interface{}{
			"A.A": "foo",
			"A.B": 42,
		})
	})
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
	t.Run("interface", func(t *testing.T) {
		a := assertions.New(t)

		v := struct {
			A interface{}
		}{
			map[string]interface{}{
				"A": "foo",
				"B": 42,
			},
		}
		m, err := MarshalByteMap(v)
		a.So(err, should.BeNil)
		a.So(m, should.Resemble, map[string][]byte{
			"A.A": mustToBytes("foo"),
			"A.B": mustToBytes(42),
		})
	})
}
