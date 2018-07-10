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

package gogoproto_test

import (
	"fmt"
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func ExampleGoFieldsPaths() {
	type cityDetails struct {
		Name string `protobuf:"name=name_city"`
	}

	type place struct {
		NameOfTheRegion string `protobuf:"name=name_region"`

		CityDetails cityDetails `protobuf:"bytes,name=city"`
	}

	london := place{
		CityDetails: cityDetails{Name: "London"},
	}
	holland := place{
		NameOfTheRegion: "Holland",
	}

	fields := gogoproto.GoFieldsPaths(&pbtypes.FieldMask{
		Paths: []string{"city.name_city"},
	}, london)
	fmt.Println(fields)

	fields = gogoproto.GoFieldsPaths(&pbtypes.FieldMask{
		Paths: []string{"name_region"},
	}, holland)
	fmt.Println(fields)

	// Output: [CityDetails.Name]
	// [NameOfTheRegion]
}

func TestGoFieldsPaths(t *testing.T) {
	a := assertions.New(t)

	type cityDetails struct {
		Name string `protobuf:"name=name_city"`
	}

	type hasProtoRenaming struct {
		NameOfTheRegion string `protobuf:"name=name_region"`

		CityDetails cityDetails `protobuf:"bytes,name=city"`
	}

	for _, tc := range []struct {
		fields, expected []string
	}{
		{
			fields:   []string{"name_region", "name_city"},
			expected: []string{"NameOfTheRegion", "name_city"},
		},
		{
			fields:   []string{"name_region"},
			expected: []string{"NameOfTheRegion"},
		},
		{
			fields:   []string{"city.name_city"},
			expected: []string{"CityDetails.Name"},
		},
	} {
		goFields := gogoproto.GoFieldsPaths(&pbtypes.FieldMask{Paths: tc.fields}, hasProtoRenaming{
			NameOfTheRegion: "england",
			CityDetails: cityDetails{
				Name: "london",
			},
		})

		a.So(goFields, should.HaveSameElementsDeep, tc.expected)
	}
}

func TestGoFieldsPathsEndDevice(t *testing.T) {
	a := assertions.New(t)

	for _, tc := range []struct {
		fields, expected []string
	}{
		{
			fields:   []string{"mac_state", "recent_uplinks", "frequency_plan_id"},
			expected: []string{"MACState", "RecentUplinks", "FrequencyPlanID"},
		},
		{
			fields:   []string{"location.latitude"},
			expected: []string{"Location.Latitude"},
		},
		{
			fields:   []string{"ids.application_ids.application_id"},
			expected: []string{"EndDeviceIdentifiers.ApplicationIdentifiers.ApplicationID"},
		},
	} {
		goFields := gogoproto.GoFieldsPaths(&pbtypes.FieldMask{Paths: tc.fields}, ttnpb.EndDevice{
			Location: &ttnpb.Location{
				Latitude: 5,
			},
		})

		a.So(goFields, should.HaveSameElementsDeep, tc.expected)
	}
}
