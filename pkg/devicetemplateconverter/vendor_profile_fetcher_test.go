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

package devicetemplateconverter_test

import (
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	mockdr "go.thethings.network/lorawan-stack/v3/pkg/devicerepository/mock"
	. "go.thethings.network/lorawan-stack/v3/pkg/devicetemplateconverter"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func Test_VendorIDProfileFetcher_ShouldFetchProfile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		device *ttnpb.EndDevice
		want   bool
	}{
		{
			name:   "Invalid",
			device: &ttnpb.EndDevice{},
			want:   false,
		},
		{
			name: "Valid",
			device: &ttnpb.EndDevice{
				VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
					VendorId: 10,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assertions.New(t)
			fetcher := &VendorIDProfileFetcher{}
			a.So(fetcher.ShouldFetchProfile(tt.device), should.Equal, tt.want)
		})
	}
}

func Test_VendorIDProfileFetcher_FetchProfile(t *testing.T) {
	t.Parallel()
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	t.Cleanup(func() {
		cancelCtx()
	})

	errNoVendorID := errors.DefineInvalidArgument("no-vendor-id", "no vendor id")

	tests := []struct {
		name         string
		endDevice    *ttnpb.EndDevice
		populateMock func(*mockdr.MockDR)
		validateResp func(*assertions.Assertion, *ttnpb.EndDeviceTemplate)
		validateErr  func(error) bool
	}{
		{
			name: "fail/no vendor id",
			endDevice: &ttnpb.EndDevice{
				VersionIds: &ttnpb.EndDeviceVersionIdentifiers{},
			},
			populateMock: func(md *mockdr.MockDR) {
				md.Err = errNoVendorID.New()
			},
			validateErr: func(err error) bool {
				return errors.IsInvalidArgument(err)
			},
		},
		{
			name: "valid",
			endDevice: &ttnpb.EndDevice{
				VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
					VendorId:        1,
					VendorProfileId: 1,
				},
			},
			populateMock: func(md *mockdr.MockDR) {
				md.EndDeviceTemplate = &ttnpb.EndDeviceTemplate{
					EndDevice: &ttnpb.EndDevice{
						VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
							VendorId:        1,
							VendorProfileId: 1,
						},
					},
					FieldMask: ttnpb.FieldMask(
						"version_ids.vendor_id",
						"version_ids.vendor_profile_id",
					),
				}
			},
			validateResp: func(a *assertions.Assertion, tmpl *ttnpb.EndDeviceTemplate) {
				// validates if there is a mock path and all the mocked Identifiers.
				a.So(ttnpb.RequireFields(tmpl.GetFieldMask().GetPaths(),
					"version_ids.vendor_id",
					"version_ids.vendor_profile_id",
				), should.BeNil)
				a.So(
					tmpl.GetEndDevice().GetVersionIds().GetVendorId() == 1 &&
						tmpl.GetEndDevice().GetVersionIds().GetVendorProfileId() == 1,
					should.BeTrue,
				)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assertions.New(t)
			drMock := mockdr.New()
			tt.populateMock(drMock)
			pf := &VendorIDProfileFetcher{
				DeviceRepository: drMock,
			}
			tmpl, err := pf.FetchProfile(ctx, tt.endDevice)
			if tt.validateErr == nil {
				a.So(err, should.BeNil)
			} else {
				a.So(tt.validateErr(err), should.BeTrue)
			}
			if tt.validateResp == nil {
				a.So(tmpl, should.BeNil)
			} else {
				tt.validateResp(a, tmpl)
			}
		})
	}
}
