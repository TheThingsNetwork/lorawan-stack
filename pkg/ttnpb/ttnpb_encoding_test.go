// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package ttnpb_test

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestStringers(t *testing.T) {
	for _, tc := range []struct {
		Stringer fmt.Stringer
		String   string
	}{
		{
			Stringer: MAC_V1_0,
			String:   "1.0.0",
		},
		{
			Stringer: MAC_V1_0_1,
			String:   "1.0.1",
		},
		{
			Stringer: MAC_V1_0_2,
			String:   "1.0.2",
		},
		{
			Stringer: MAC_V1_1,
			String:   "1.1.0",
		},
		{
			Stringer: PHY_V1_0,
			String:   "1.0.0",
		},
		{
			Stringer: PHY_V1_0_1,
			String:   "1.0.1",
		},
		{
			Stringer: PHY_V1_0_2_REV_A,
			String:   "1.0.2-a",
		},
		{
			Stringer: PHY_V1_0_2_REV_B,
			String:   "1.0.2-b",
		},
		{
			Stringer: PHY_V1_1_REV_A,
			String:   "1.1.0-a",
		},
		{
			Stringer: PHY_V1_1_REV_B,
			String:   "1.1.0-b",
		},
	} {
		assertions.New(t).So(tc.Stringer.String(), should.Equal, tc.String)
	}
}

func TestEnumMarshalers(t *testing.T) {
	var vals [][]fmt.Stringer

	var macVers []fmt.Stringer
	for i := range MACVersion_name {
		macVers = append(macVers, MACVersion(i))
	}
	vals = append(vals, macVers)

	var phyVers []fmt.Stringer
	for i := range PHYVersion_name {
		phyVers = append(phyVers, PHYVersion(i))
	}
	vals = append(vals, phyVers)

	var grants []fmt.Stringer
	for i := range GrantType_name {
		grants = append(grants, GrantType(i))
	}
	vals = append(vals, grants)

	var clusterRoles []fmt.Stringer
	for i := range ClusterRole_name {
		clusterRoles = append(clusterRoles, ClusterRole(i))
	}
	vals = append(vals, clusterRoles)

	var states []fmt.Stringer
	for i := range State_name {
		states = append(states, State(i))
	}
	vals = append(vals, states)

	var locationSources []fmt.Stringer
	for i := range LocationSource_name {
		locationSources = append(locationSources, LocationSource(i))
	}
	vals = append(vals, locationSources)

	var rights []fmt.Stringer
	for i := range Right_name {
		rights = append(rights, Right(i))
	}
	vals = append(vals, rights)

	for _, vs := range vals {
		typ := reflect.TypeOf(vs[0])
		newV := func() interface{} { return reflect.New(typ).Interface() }

		t.Run(typ.String(), func(t *testing.T) {
			for _, v := range vs {
				a := assertions.New(t)
				if !a.So(func() { _ = v.String() }, should.NotPanic) {
					t.FailNow()
				}

				t.Run(v.String(), func(t *testing.T) {
					if m, ok := v.(encoding.TextMarshaler); ok {
						t.Run("Text", func(t *testing.T) {
							b, err := m.MarshalText()
							a.So(err, should.BeNil)

							got, ok := newV().(encoding.TextUnmarshaler)
							if !ok {
								t.Fatal("Does not implement TextUnmarshaler")
							}

							err = got.UnmarshalText(b)
							a.So(err, should.BeNil)
							a.So(reflect.Indirect(reflect.ValueOf(got)).Interface(), should.Resemble, v)

							got = newV().(encoding.TextUnmarshaler)
							err = got.UnmarshalText([]byte(v.(fmt.Stringer).String()))
							a.So(err, should.BeNil)
							a.So(reflect.Indirect(reflect.ValueOf(got)).Interface(), should.Resemble, v)
						})
					}

					if m, ok := v.(encoding.BinaryMarshaler); ok {
						t.Run("Binary", func(t *testing.T) {
							b, err := m.MarshalBinary()
							a.So(err, should.BeNil)

							got, ok := newV().(encoding.BinaryUnmarshaler)
							if !ok {
								t.Fatal("Does not implement BinaryUnmarshaler")
							}

							err = got.UnmarshalBinary(b)
							a.So(err, should.BeNil)
							a.So(reflect.Indirect(reflect.ValueOf(got)).Interface(), should.Resemble, v)
						})
					}

					if m, ok := v.(json.Marshaler); ok {
						t.Run("JSON", func(t *testing.T) {
							b, err := m.MarshalJSON()
							a.So(err, should.BeNil)

							got, ok := newV().(json.Unmarshaler)
							if !ok {
								t.Fatal("Does not implement JSONUnmarshaler")
							}

							err = got.UnmarshalJSON(b)
							a.So(err, should.BeNil)
							a.So(reflect.Indirect(reflect.ValueOf(got)).Interface(), should.Resemble, v)
						})
					}
				})
			}
		})
	}
}
