// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package normalizedpayload_test

import (
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors/normalizedpayload"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/structpb"
)

func timePtr(t time.Time) *time.Time {
	return &t
}

func float64Ptr(f float64) *float64 {
	return &f
}

func TestUplink(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name                     string
		normalizedPayload        []*structpb.Struct
		expected                 []normalizedpayload.Measurement
		expectedValidationErrors [][]error
		errorAssertion           func(error) bool
	}{
		{
			name: "single timestamp",
			normalizedPayload: []*structpb.Struct{
				{
					Fields: map[string]*structpb.Value{
						"time": {
							Kind: &structpb.Value_StringValue{
								StringValue: "2022-08-23T17:13:42Z",
							},
						},
					},
				},
			},
			expected: []normalizedpayload.Measurement{
				{
					Time: timePtr(time.Date(2022, 8, 23, 17, 13, 42, 0, time.UTC)),
				},
			},
		},
		{
			name: "one soil nutrient concentration",
			normalizedPayload: []*structpb.Struct{
				{
					Fields: map[string]*structpb.Value{
						"soil": {
							Kind: &structpb.Value_StructValue{
								StructValue: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"n": {
											Kind: &structpb.Value_NumberValue{
												NumberValue: 999999.99,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []normalizedpayload.Measurement{
				{
					Soil: normalizedpayload.Soil{
						Nitrogen: float64Ptr(999999.99),
					},
				},
			},
		},
		{
			name: "two air temperatures",
			normalizedPayload: []*structpb.Struct{
				{
					Fields: map[string]*structpb.Value{
						"air": {
							Kind: &structpb.Value_StructValue{
								StructValue: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"temperature": {
											Kind: &structpb.Value_NumberValue{
												NumberValue: 20.42,
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Fields: map[string]*structpb.Value{
						"air": {
							Kind: &structpb.Value_StructValue{
								StructValue: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"temperature": {
											Kind: &structpb.Value_NumberValue{
												NumberValue: 19.61,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []normalizedpayload.Measurement{
				{
					Air: normalizedpayload.Air{
						Temperature: float64Ptr(20.42),
					},
				},
				{
					Air: normalizedpayload.Air{
						Temperature: float64Ptr(19.61),
					},
				},
			},
		},
		{
			name: "no fields",
			normalizedPayload: []*structpb.Struct{
				{},
			},
			expected: []normalizedpayload.Measurement{
				{},
			},
		},
		{
			name: "above 100 percent soil moisture",
			normalizedPayload: []*structpb.Struct{
				{
					Fields: map[string]*structpb.Value{
						"soil": {
							Kind: &structpb.Value_StructValue{
								StructValue: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"moisture": {
											Kind: &structpb.Value_NumberValue{
												NumberValue: 120,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []normalizedpayload.Measurement{
				{},
			},
			expectedValidationErrors: [][]error{
				{
					normalizedpayload.ErrFieldMaximum.WithAttributes(
						"path", "soil.moisture",
						"maximum", 100.0,
					),
				},
			},
		},
		{
			name: "below absolute zero",
			normalizedPayload: []*structpb.Struct{
				{
					Fields: map[string]*structpb.Value{
						"air": {
							Kind: &structpb.Value_StructValue{
								StructValue: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"temperature": {
											Kind: &structpb.Value_NumberValue{
												NumberValue: -300,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []normalizedpayload.Measurement{
				{},
			},
			expectedValidationErrors: [][]error{
				{
					normalizedpayload.ErrFieldMinimum.WithAttributes(
						"path", "air.temperature",
						"minimum", -273.15,
					),
				},
			},
		},
		{
			name: "invalid direction",
			normalizedPayload: []*structpb.Struct{
				{
					Fields: map[string]*structpb.Value{
						"wind": {
							Kind: &structpb.Value_StructValue{
								StructValue: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"direction": {
											Kind: &structpb.Value_NumberValue{
												NumberValue: 360, // this is 0 degrees
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []normalizedpayload.Measurement{
				{},
			},
			expectedValidationErrors: [][]error{
				{
					normalizedpayload.ErrFieldExclusiveMaximum.WithAttributes(
						"path", "wind.direction",
						"maximum", 360.0,
					),
				},
			},
		},
		{
			name: "invalid type",
			normalizedPayload: []*structpb.Struct{
				{
					Fields: map[string]*structpb.Value{
						"air": {
							Kind: &structpb.Value_StructValue{
								StructValue: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"temperature": {
											Kind: &structpb.Value_StringValue{
												StringValue: "test",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			errorAssertion: errors.IsInvalidArgument,
		},
		{
			name: "unknown field",
			normalizedPayload: []*structpb.Struct{
				{
					Fields: map[string]*structpb.Value{
						"air": {
							Kind: &structpb.Value_StructValue{
								StructValue: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"unknown": {
											Kind: &structpb.Value_StringValue{
												StringValue: "test",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			errorAssertion: errors.IsInvalidArgument,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assertions.New(t)

			measurements, err := normalizedpayload.Parse(tc.normalizedPayload)
			if tc.errorAssertion != nil {
				a.So(err, should.NotBeNil)
				a.So(tc.errorAssertion(err), should.BeTrue)
			} else {
				a.So(err, should.BeNil)
				if !a.So(measurements, should.HaveLength, len(tc.expected)) {
					t.FailNow()
				}
				for i, parsed := range measurements {
					if len(parsed.ValidationErrors) > 0 {
						a.So(len(tc.expectedValidationErrors), should.BeGreaterThanOrEqualTo, i+1)
						a.So(parsed.ValidationErrors, should.HaveLength, len(tc.expectedValidationErrors[i]))
						for j, err := range parsed.ValidationErrors {
							a.So(err, should.EqualErrorOrDefinition, tc.expectedValidationErrors[i][j])
						}
					}
					a.So(parsed.Measurement, should.Resemble, tc.expected[i])
				}
			}
		})
	}
}
