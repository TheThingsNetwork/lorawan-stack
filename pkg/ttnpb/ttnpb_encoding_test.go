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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
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

	var mTypes []fmt.Stringer
	for i := range MType_name {
		mTypes = append(mTypes, MType(i))
	}
	vals = append(vals, mTypes)

	var majors []fmt.Stringer
	for i := range Major_name {
		majors = append(majors, Major(i))
	}
	vals = append(vals, majors)

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

	var drIdxs []fmt.Stringer
	for i := range DataRateIndex_name {
		drIdxs = append(drIdxs, DataRateIndex(i))
	}
	vals = append(vals, drIdxs)

	var drOffsets []fmt.Stringer
	for i := range DataRateOffset_name {
		drOffsets = append(drOffsets, DataRateOffset(i))
	}
	vals = append(vals, drOffsets)

	var rejoins []fmt.Stringer
	for i := range RejoinType_name {
		rejoins = append(rejoins, RejoinType(i))
	}
	vals = append(vals, rejoins)

	var cfLists []fmt.Stringer
	for i := range CFListType_name {
		cfLists = append(cfLists, CFListType(i))
	}
	vals = append(vals, cfLists)

	var classes []fmt.Stringer
	for i := range Class_name {
		classes = append(classes, Class(i))
	}
	vals = append(vals, classes)

	var txSchedulePrios []fmt.Stringer
	for i := range TxSchedulePriority_name {
		txSchedulePrios = append(txSchedulePrios, TxSchedulePriority(i))
	}
	vals = append(vals, txSchedulePrios)

	var cids []fmt.Stringer
	for i := range MACCommandIdentifier_name {
		cids = append(cids, MACCommandIdentifier(i))
	}
	vals = append(vals, cids)

	var dutyCycles []fmt.Stringer
	for i := range AggregatedDutyCycle_name {
		dutyCycles = append(dutyCycles, AggregatedDutyCycle(i))
	}
	vals = append(vals, dutyCycles)

	var pingSlots []fmt.Stringer
	for i := range PingSlotPeriod_name {
		pingSlots = append(pingSlots, PingSlotPeriod(i))
	}
	vals = append(vals, pingSlots)

	var rejoinCounts []fmt.Stringer
	for i := range RejoinCountExponent_name {
		rejoinCounts = append(rejoinCounts, RejoinCountExponent(i))
	}
	vals = append(vals, rejoinCounts)

	var rejoinTimes []fmt.Stringer
	for i := range RejoinTimeExponent_name {
		rejoinTimes = append(rejoinTimes, RejoinTimeExponent(i))
	}
	vals = append(vals, rejoinTimes)

	var rejoinPeriods []fmt.Stringer
	for i := range RejoinPeriodExponent_name {
		rejoinPeriods = append(rejoinPeriods, RejoinPeriodExponent(i))
	}
	vals = append(vals, rejoinPeriods)

	var deviceEIRPs []fmt.Stringer
	for i := range DeviceEIRP_name {
		deviceEIRPs = append(deviceEIRPs, DeviceEIRP(i))
	}
	vals = append(vals, deviceEIRPs)

	var ackLimitExponents []fmt.Stringer
	for i := range ADRAckLimitExponent_name {
		ackLimitExponents = append(ackLimitExponents, ADRAckLimitExponent(i))
	}
	vals = append(vals, ackLimitExponents)

	var ackDelayExponents []fmt.Stringer
	for i := range ADRAckDelayExponent_name {
		ackDelayExponents = append(ackDelayExponents, ADRAckDelayExponent(i))
	}
	vals = append(vals, ackDelayExponents)

	var rxDelays []fmt.Stringer
	for i := range RxDelay_name {
		rxDelays = append(rxDelays, RxDelay(i))
	}
	vals = append(vals, rxDelays)

	var minors []fmt.Stringer
	for i := range Minor_name {
		minors = append(minors, Minor(i))
	}
	vals = append(vals, minors)

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

	var outLines []string
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
							a := assertions.New(t)
							b, err := m.MarshalText()
							if !a.So(err, should.BeNil) {
								t.Error(test.FormatError(err))
							}
							outLines = append(outLines, fmt.Sprintf(`Text: %s "%s" -> "%s"`, typ, v, b))

							got, ok := newV().(encoding.TextUnmarshaler)
							if !ok {
								t.Fatal("Does not implement TextUnmarshaler")
							}

							err = got.UnmarshalText(b)
							a.So(err, should.BeNil)
							a.So(reflect.Indirect(reflect.ValueOf(got)).Interface(), should.Resemble, v)

							got = newV().(encoding.TextUnmarshaler)
							err = got.UnmarshalText([]byte(v.(fmt.Stringer).String()))
							if !a.So(err, should.BeNil) {
								t.Error(test.FormatError(err))
							}
							a.So(reflect.Indirect(reflect.ValueOf(got)).Interface(), should.Resemble, v)
						})
					}

					if m, ok := v.(encoding.BinaryMarshaler); ok {
						t.Run("Binary", func(t *testing.T) {
							a := assertions.New(t)
							b, err := m.MarshalBinary()
							if !a.So(err, should.BeNil) {
								t.Error(test.FormatError(err))
							}
							outLines = append(outLines, fmt.Sprintf(`Binary: %s "%s" -> %v`, typ, v, b))

							got, ok := newV().(encoding.BinaryUnmarshaler)
							if !ok {
								t.Fatal("Does not implement BinaryUnmarshaler")
							}

							err = got.UnmarshalBinary(b)
							if !a.So(err, should.BeNil) {
								t.Error(test.FormatError(err))
							}
							a.So(reflect.Indirect(reflect.ValueOf(got)).Interface(), should.Resemble, v)
						})
					}

					if m, ok := v.(json.Marshaler); ok {
						t.Run("JSON", func(t *testing.T) {
							a := assertions.New(t)
							b, err := m.MarshalJSON()
							if !a.So(err, should.BeNil) {
								t.Error(test.FormatError(err))
							}
							outLines = append(outLines, fmt.Sprintf(`JSON: %s "%s" -> "%s"`, typ, v, b))

							got, ok := newV().(json.Unmarshaler)
							if !ok {
								t.Fatal("Does not implement JSONUnmarshaler")
							}

							err = got.UnmarshalJSON(b)
							if !a.So(err, should.BeNil) {
								t.Error(test.FormatError(err))
							}
							a.So(reflect.Indirect(reflect.ValueOf(got)).Interface(), should.Resemble, v)
						})
					}
				})
			}
		})
	}

	if t.Failed() {
		return
	}
	sort.Strings(outLines)
	out := strings.Join(outLines, "\n")
	goldenPath := filepath.Join("testdata", "ttnpb_encoding_golden")
	if os.Getenv("TEST_WRITE_GOLDEN") == "1" {
		if err := ioutil.WriteFile(goldenPath, []byte(out), 0o644); err != nil {
			t.Fatalf("Failed to write golden file: %s", err)
		}
	} else {
		prevOut, err := ioutil.ReadFile(goldenPath)
		if err != nil {
			t.Fatalf("Failed to read golden file: %s", err)
		}
		assertions.New(t).So(out, should.Resemble, string(prevOut))
	}
}
