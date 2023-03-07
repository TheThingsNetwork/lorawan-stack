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

package alcsyncv1

import (
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestPackageDataExtractsStructCorrectly(t *testing.T) {
	t.Parallel()
	a, _ := test.New(t)
	expected := &packageData{
		FPort:     202,
		Threshold: time.Duration(10) * time.Second,
	}
	st := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"f_port": {
				Kind: &structpb.Value_NumberValue{
					NumberValue: float64(expected.FPort),
				},
			},
			"threshold": {
				Kind: &structpb.Value_NumberValue{
					NumberValue: 10,
				},
			},
		},
	}
	actual := &packageData{}
	err := actual.fromStruct(st)
	a.So(err, should.BeNil)
	a.So(actual, should.Resemble, expected)
}

func TestPackageDataHandlesInvalidValues(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string
		St   *structpb.Struct
	}{
		{
			Name: "InvalidThreshold",
			St: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"threshold": {
						Kind: &structpb.Value_StringValue{
							StringValue: "10s",
						},
					},
				},
			},
		},
		{
			Name: "InvalidFPort",
			St: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"f_port": {
						Kind: &structpb.Value_StringValue{
							StringValue: "202",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a, _ := test.New(t)
			actual := &packageData{}
			err := actual.fromStruct(tc.St)
			a.So(err, should.Resemble, errInvalidFieldType.New())
		})
	}
}

func TestPackageDataMerge(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string
		In   struct {
			DefaultAssoc *ttnpb.ApplicationPackageDefaultAssociation
			PkgAssoc     *ttnpb.ApplicationPackageAssociation
		}
		Expected *packageData
	}{
		{
			Name: "PopulatesValuesFromDefaultAssocOnly",
			In: struct {
				DefaultAssoc *ttnpb.ApplicationPackageDefaultAssociation
				PkgAssoc     *ttnpb.ApplicationPackageAssociation
			}{
				DefaultAssoc: &ttnpb.ApplicationPackageDefaultAssociation{
					Data: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"f_port": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 202,
								},
							},
							"threshold": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 4,
								},
							},
						},
					},
				},
			},
			Expected: &packageData{
				FPort:     202,
				Threshold: time.Duration(4) * time.Second,
			},
		},
		{
			Name: "PopulatesValuesFromPkgAssocOnly",
			In: struct {
				DefaultAssoc *ttnpb.ApplicationPackageDefaultAssociation
				PkgAssoc     *ttnpb.ApplicationPackageAssociation
			}{
				PkgAssoc: &ttnpb.ApplicationPackageAssociation{
					Data: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"f_port": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 202,
								},
							},
							"threshold": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 4,
								},
							},
						},
					},
				},
			},
			Expected: &packageData{
				FPort:     202,
				Threshold: time.Duration(4) * time.Second,
			},
		},
		{
			Name: "PopulatesValuesFromPkgAssocOverridingDefaultAssoc",
			In: struct {
				DefaultAssoc *ttnpb.ApplicationPackageDefaultAssociation
				PkgAssoc     *ttnpb.ApplicationPackageAssociation
			}{
				DefaultAssoc: &ttnpb.ApplicationPackageDefaultAssociation{
					Data: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"f_port": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 202,
								},
							},
							"threshold": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 4,
								},
							},
						},
					},
				},
				PkgAssoc: &ttnpb.ApplicationPackageAssociation{
					Data: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"f_port": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 203,
								},
							},
							"threshold": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 5,
								},
							},
						},
					},
				},
			},
			Expected: &packageData{
				FPort:     203,
				Threshold: time.Duration(5) * time.Second,
			},
		},
		{
			Name: "PopulatesDefaultValueForThreshold",
			In: struct {
				DefaultAssoc *ttnpb.ApplicationPackageDefaultAssociation
				PkgAssoc     *ttnpb.ApplicationPackageAssociation
			}{
				DefaultAssoc: &ttnpb.ApplicationPackageDefaultAssociation{
					Data: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"f_port": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 202,
								},
							},
						},
					},
				},
			},
			Expected: &packageData{
				FPort:     202,
				Threshold: defaultThreshold,
			},
		},
		{
			Name: "PopulatesDefaultValueForFPortFromDefaultAssocIds",
			In: struct {
				DefaultAssoc *ttnpb.ApplicationPackageDefaultAssociation
				PkgAssoc     *ttnpb.ApplicationPackageAssociation
			}{
				DefaultAssoc: &ttnpb.ApplicationPackageDefaultAssociation{
					Ids: &ttnpb.ApplicationPackageDefaultAssociationIdentifiers{
						FPort: 218,
					},
				},
			},
			Expected: &packageData{
				FPort:     218,
				Threshold: defaultThreshold,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a, _ := test.New(t)
			actual, err := mergePackageData(tc.In.DefaultAssoc, tc.In.PkgAssoc)
			a.So(err, should.BeNil)
			a.So(actual, should.Resemble, tc.Expected)
		})
	}
}
