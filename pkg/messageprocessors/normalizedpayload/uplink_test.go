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

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors/normalizedpayload"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
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
		name              string
		normalizedPayload []*pbtypes.Struct
		expected          []normalizedpayload.Measurement
		errorAssertion    func(error) bool
	}{
		{
			name: "single timestamp",
			normalizedPayload: []*pbtypes.Struct{
				{
					Fields: map[string]*pbtypes.Value{
						"time": {
							Kind: &pbtypes.Value_StringValue{
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
			name: "two air temperatures",
			normalizedPayload: []*pbtypes.Struct{
				{
					Fields: map[string]*pbtypes.Value{
						"air": {
							Kind: &pbtypes.Value_StructValue{
								StructValue: &pbtypes.Struct{
									Fields: map[string]*pbtypes.Value{
										"temperature": {
											Kind: &pbtypes.Value_NumberValue{
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
					Fields: map[string]*pbtypes.Value{
						"air": {
							Kind: &pbtypes.Value_StructValue{
								StructValue: &pbtypes.Struct{
									Fields: map[string]*pbtypes.Value{
										"temperature": {
											Kind: &pbtypes.Value_NumberValue{
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
					Air: &normalizedpayload.Air{
						Temperature: float64Ptr(20.42),
					},
				},
				{
					Air: &normalizedpayload.Air{
						Temperature: float64Ptr(19.61),
					},
				},
			},
		},
		{
			name: "below absolute zero",
			normalizedPayload: []*pbtypes.Struct{
				{
					Fields: map[string]*pbtypes.Value{
						"air": {
							Kind: &pbtypes.Value_StructValue{
								StructValue: &pbtypes.Struct{
									Fields: map[string]*pbtypes.Value{
										"temperature": {
											Kind: &pbtypes.Value_NumberValue{
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
			errorAssertion: errors.IsInvalidArgument,
		},
		{
			name: "invalid direction",
			normalizedPayload: []*pbtypes.Struct{
				{
					Fields: map[string]*pbtypes.Value{
						"wind": {
							Kind: &pbtypes.Value_StructValue{
								StructValue: &pbtypes.Struct{
									Fields: map[string]*pbtypes.Value{
										"direction": {
											Kind: &pbtypes.Value_NumberValue{
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
			errorAssertion: errors.IsInvalidArgument,
		},
		{
			name: "invalid type",
			normalizedPayload: []*pbtypes.Struct{
				{
					Fields: map[string]*pbtypes.Value{
						"air": {
							Kind: &pbtypes.Value_StructValue{
								StructValue: &pbtypes.Struct{
									Fields: map[string]*pbtypes.Value{
										"temperature": {
											Kind: &pbtypes.Value_StringValue{
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
			normalizedPayload: []*pbtypes.Struct{
				{
					Fields: map[string]*pbtypes.Value{
						"air": {
							Kind: &pbtypes.Value_StructValue{
								StructValue: &pbtypes.Struct{
									Fields: map[string]*pbtypes.Value{
										"unknown": {
											Kind: &pbtypes.Value_StringValue{
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
				a.So(measurements, should.Resemble, tc.expected)
			}
		})
	}
}
