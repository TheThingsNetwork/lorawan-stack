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

package store_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

type cityDetails struct {
	Name string `protobuf:"name=name_city"`
}

type hasProtoRenaming struct {
	NameOfTheRegion string `protobuf:"name=name_region"`

	CityDetails cityDetails `protobuf:"bytes,name=city"`
}

type hasNoProtoRenaming struct {
	Region string
	City   string
}

func TestConvertProtoFields(t *testing.T) {
	entry := hasProtoRenaming{
		NameOfTheRegion: "england",
		CityDetails: cityDetails{
			Name: "london",
		},
	}

	for i, fieldsCase := range []struct {
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
		goFields := store.ConvertProtoFields(fieldsCase.fields, reflect.ValueOf(entry))

	fields:
		for _, expectedField := range fieldsCase.expected {
			for _, foundGoField := range goFields {
				if foundGoField == expectedField {
					continue fields
				}
			}

			t.Fatalf("Case %d: Did not find expected field `%s`, found instead `%s`\n",
				i+1, expectedField, strings.Join(goFields, "`, `"))
		}
	}
}

func TestConvertProtoFieldsEndDevice(t *testing.T) {
	dev := ttnpb.EndDevice{
		Location: &ttnpb.Location{
			Latitude: 5,
		},
	}

	for i, fieldsCase := range []struct {
		fields, expected []string
	}{
		{
			fields:   []string{"mac_state_desired", "lorawan_version"},
			expected: []string{"MACStateDesired", "LoRaWANVersion"},
		},
		{
			fields:   []string{"location.latitude"},
			expected: []string{"Location.Latitude"},
		},
	} {
		goFields := store.ConvertProtoFields(fieldsCase.fields, reflect.ValueOf(dev))

	fields:
		for _, expectedField := range fieldsCase.expected {
			for _, foundGoField := range goFields {
				if foundGoField == expectedField {
					continue fields
				}
			}

			t.Fatalf("Case %d: Did not find expected field `%s`, found instead `%s`\n",
				i+1, expectedField, strings.Join(goFields, "`, `"))
		}
	}
}
