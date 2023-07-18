// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package loracloudgeolocationv3

import (
	"math/rand"
	"net/url"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loragls/v3/api"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/structpb"
)

// nolint: gosec
func generateRandomizedMD(count int) [][]*api.RxMetadata {
	md := make([][]*api.RxMetadata, count)
	for i := range md {
		md[i] = make([]*api.RxMetadata, rand.Intn(10))
		for j := range md[i] {
			md[i][j] = &api.RxMetadata{
				GatewayIDs: &api.GatewayIDs{
					GatewayID: "test-gateway-id",
				},
				AntennaIndex:  rand.Uint32(),
				FineTimestamp: rand.Uint64(),
				RSSI:          rand.Float32(),
				SNR:           rand.Float32(),
				Location: &api.RxMDLocation{
					Latitude:  rand.Float64()*180 - 90,
					Longitude: rand.Float64()*360 - 180,
					Altitude:  rand.Int31(),
					Accuracy:  rand.Int31(),
				},
			}
		}
	}
	return md
}

func queryTypePtr(i uint8) *QueryType {
	qt := QueryType(i)
	return &qt
}

func boolPtr(b bool) *bool {
	return &b
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}

func intPtr(i int) *int {
	return &i
}

func stringPtr(str string) *string {
	return &str
}

func mustParse(str string) *url.URL {
	u, err := url.Parse(str)
	if err != nil {
		panic(err)
	}
	return u
}

func TestRecentMetadataRoundtrip(t *testing.T) {
	t.Parallel()
	a, _ := test.New(t)
	now := time.Now().UTC().Truncate(time.Second)

	rxMDs := generateRandomizedMD(5)
	uplinkMDs := make([]*UplinkMetadata, len(rxMDs))
	for i, rxMD := range rxMDs {
		uplinkMDs[i] = &UplinkMetadata{
			RxMetadata: rxMD,
			ReceivedAt: now.Add(time.Duration(i) * time.Second),
		}
	}
	expectedData := Data{
		Query:                queryTypePtr(1),
		MultiFrame:           boolPtr(true),
		MultiFrameWindowSize: intPtr(10),
		MultiFrameWindowAge:  durationPtr(10 * time.Minute),
		ServerURL:            mustParse("https://example.com"),
		Token:                stringPtr("test-token"),
		RecentMetadata:       uplinkMDs,
	}
	st, err := expectedData.Struct()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	actualData := Data{}
	if err := actualData.FromStruct(st); !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(actualData, should.Resemble, expectedData)
}

func TestPackageDataDeserialization(t *testing.T) {
	t.Parallel()
	a, _ := test.New(t)

	testCases := []struct {
		Name          string
		InputData     map[string]any
		ExpectedData  *Data
		ExpectedError error
	}{
		{
			Name:          "DoesNotDeserializeEmptyData",
			InputData:     map[string]any{},
			ExpectedData:  &Data{},
			ExpectedError: nil,
		},
		{
			Name: "SuccessfullyDeserializesData",
			InputData: map[string]any{
				queryField:           "TOARSSI",
				multiFrameField:      true,
				multiFrameWindowSize: 10,
				multiFrameWindowAge:  10,
				serverURLField:       "https://example.com",
				tokenField:           "test-token",
			},
			ExpectedData: &Data{
				Query:                queryTypePtr(1),
				MultiFrame:           boolPtr(true),
				MultiFrameWindowSize: intPtr(10),
				MultiFrameWindowAge:  durationPtr(10 * time.Minute),
				ServerURL:            mustParse("https://example.com"),
				Token:                stringPtr("test-token"),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			data := &Data{}
			st, err := structpb.NewStruct(tc.InputData)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			err = data.FromStruct(st)
			a.So(err, should.EqualErrorOrDefinition, tc.ExpectedError)
		})
	}
}

func TestPackageDataSerialization(t *testing.T) {
	t.Parallel()
	a, _ := test.New(t)

	testCases := []struct {
		Name          string
		InputData     *Data
		ExpectedData  *structpb.Struct
		ExpectedError error
	}{
		{
			Name:          "DoesNotSerializezNilValues",
			InputData:     &Data{},
			ExpectedError: nil,
			ExpectedData:  nil,
		},
		{
			Name: "DoesNotSerializeEmptyValues",
			InputData: &Data{
				Query:          nil,
				ServerURL:      &url.URL{},
				Token:          new(string),
				RecentMetadata: []*UplinkMetadata{},
			},
			ExpectedError: nil,
			ExpectedData:  nil,
		},
		{
			Name: "SerializesData",
			InputData: &Data{
				Query:                queryTypePtr(0),
				MultiFrame:           boolPtr(true),
				MultiFrameWindowSize: intPtr(10),
				MultiFrameWindowAge:  durationPtr(10 * time.Minute),
				ServerURL:            mustParse("https://example.com"),
				Token:                stringPtr("test-token"),
			},
			ExpectedError: nil,
			ExpectedData: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					queryField:           structpb.NewStringValue("TOARSSI"),
					multiFrameField:      structpb.NewBoolValue(true),
					multiFrameWindowSize: structpb.NewNumberValue(10),
					multiFrameWindowAge:  structpb.NewNumberValue(10),
					serverURLField:       structpb.NewStringValue("https://example.com"),
					tokenField:           structpb.NewStringValue("test-token"),
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			actualData, err := tc.InputData.Struct()

			a.So(err, should.EqualErrorOrDefinition, tc.ExpectedError)
			a.So(actualData, should.Resemble, tc.ExpectedData)
		})
	}
}

func TestPackageDataMerge(t *testing.T) {
	t.Parallel()
	a, _ := test.New(t)

	testCases := []struct {
		Name            string
		DefaultData     Data
		AssociationData Data
		ExpectedData    *Data
	}{
		{
			Name: "TakesDefaultData",
			DefaultData: Data{
				Query:                queryTypePtr(1),
				MultiFrame:           boolPtr(true),
				MultiFrameWindowSize: intPtr(10),
				MultiFrameWindowAge:  durationPtr(10 * time.Minute),
				ServerURL:            mustParse("https://example.com"),
				Token:                stringPtr("test-token"),
			},
			AssociationData: Data{},
			ExpectedData: &Data{
				Query:                queryTypePtr(1),
				MultiFrame:           boolPtr(true),
				MultiFrameWindowSize: intPtr(10),
				MultiFrameWindowAge:  durationPtr(10 * time.Minute),
				ServerURL:            mustParse("https://example.com"),
				Token:                stringPtr("test-token"),
			},
		},
		{
			Name: "FillsEmptyValues",
			DefaultData: Data{
				Query:     queryTypePtr(1),
				ServerURL: mustParse("https://example.com"),
				Token:     stringPtr("test-token"),
			},
			AssociationData: Data{
				MultiFrame:           boolPtr(true),
				MultiFrameWindowSize: intPtr(10),
				MultiFrameWindowAge:  durationPtr(10 * time.Minute),
			},
			ExpectedData: &Data{
				Query:                queryTypePtr(1),
				MultiFrame:           boolPtr(true),
				MultiFrameWindowSize: intPtr(10),
				MultiFrameWindowAge:  durationPtr(10 * time.Minute),
				ServerURL:            mustParse("https://example.com"),
				Token:                stringPtr("test-token"),
			},
		},
		{
			Name: "OverridesDefaultValues",
			DefaultData: Data{
				Query:                queryTypePtr(1),
				MultiFrame:           boolPtr(true),
				MultiFrameWindowSize: intPtr(10),
				MultiFrameWindowAge:  durationPtr(10 * time.Minute),
				ServerURL:            mustParse("https://example.com"),
				Token:                stringPtr("test-token"),
			},
			AssociationData: Data{
				Query:                queryTypePtr(2),
				MultiFrame:           boolPtr(false),
				MultiFrameWindowSize: intPtr(4),
				MultiFrameWindowAge:  durationPtr(3 * time.Minute),
				ServerURL:            mustParse("https://other.example.com"),
				Token:                stringPtr("other-test-token"),
			},
			ExpectedData: &Data{
				Query:                queryTypePtr(2),
				MultiFrame:           boolPtr(false),
				MultiFrameWindowSize: intPtr(4),
				MultiFrameWindowAge:  durationPtr(3 * time.Minute),
				ServerURL:            mustParse("https://other.example.com"),
				Token:                stringPtr("other-test-token"),
			},
		},
		{
			Name:            "PopulatesDefaultPackageValues",
			DefaultData:     Data{},
			AssociationData: Data{},
			ExpectedData: &Data{
				ServerURL: api.DefaultServerURL,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			actualData := mergeData(tc.DefaultData, tc.AssociationData)
			a.So(actualData, should.Resemble, tc.ExpectedData)
		})
	}
}

func TestPackageDataValidation(t *testing.T) {
	t.Parallel()
	a, _ := test.New(t)
	testCases := []struct {
		Name       string
		MergedData *Data
		Error      error
	}{
		{
			Name: "ValidData",
			MergedData: &Data{
				Query:     queryTypePtr(1),
				Token:     stringPtr("test-token"),
				ServerURL: mustParse("https://example.com"),
			},
			Error: nil,
		},
		{
			Name: "MissingQuery",
			MergedData: &Data{
				Token:     stringPtr("test-token"),
				ServerURL: mustParse("https://example.com"),
			},
			Error: errFieldRequired.WithAttributes("field", queryField),
		},
		{
			Name: "MissingToken",
			MergedData: &Data{
				Query:     queryTypePtr(1),
				ServerURL: mustParse("https://example.com"),
			},
			Error: errFieldRequired.WithAttributes("field", tokenField),
		},
		{
			Name: "MissingServerURL",
			MergedData: &Data{
				Query: queryTypePtr(1),
				Token: stringPtr("test-token"),
			},
			Error: errFieldRequired.WithAttributes("field", serverURLField),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			err := validateData(tc.MergedData)
			a.So(err, should.EqualErrorOrDefinition, tc.Error)
		})
	}
}
