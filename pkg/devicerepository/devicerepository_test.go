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

package devicerepository_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/devicerepository"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	validFetcher = fetch.NewMemFetcher(map[string][]byte{
		"brands.yml": []byte(`version: '3'
brands:
  thethingsproducts:
    name: The Things Products
    url: https://www.thethingsnetwork.org
    logos:
    - logo.png`),
		"thethingsproducts/devices.yml": []byte(`version: '3'
devices:
  thethingsuno:
    name: The Things Uno`),
		"thethingsproducts/thethingsuno/versions.yml": []byte(`version: '3'
hardware_versions:
  '1.0':
    - firmware_version: 1.1
      photos: [front.jpg, back.jpg]
      payload_format:
        up:
          type: grpc
          parameter: hosted-service:1234
        down:
          type: javascript
          parameter: encoder.js`),
		"thethingsproducts/thethingsuno/1.0/encoder.js": []byte(`function Encoder() { return { led: 1 } }`)})

	invalidFetcher = fetch.NewMemFetcher(map[string][]byte{
		"brands.yml":                                  []byte(`invalid yaml`),
		"thethingsproducts/devices.yml":               []byte(`invalid yaml`),
		"thethingsproducts/thethingsuno/versions.yml": []byte(`invalid yaml`)})

	emptyFetcher = fetch.NewMemFetcher(map[string][]byte{})
)

func TestBrand(t *testing.T) {
	for _, tc := range []struct {
		Name          string
		Fetcher       fetch.Interface
		ExpectedErr   func(err error) bool
		ExpectedValue interface{}
	}{
		{
			Name:        "Normal",
			Fetcher:     validFetcher,
			ExpectedErr: func(err error) bool { return err == nil },
			ExpectedValue: map[string]ttnpb.EndDeviceBrand{
				"thethingsproducts": {
					ID:    "thethingsproducts",
					Name:  "The Things Products",
					URL:   "https://www.thethingsnetwork.org",
					Logos: []string{"logo.png"},
				},
			},
		},
		{
			Name:        "Invalid",
			Fetcher:     invalidFetcher,
			ExpectedErr: errors.IsInvalidArgument,
		},
		{
			Name:        "Empty",
			Fetcher:     emptyFetcher,
			ExpectedErr: errors.IsNotFound,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			repo := Client{Fetcher: tc.Fetcher}
			brands, err := repo.Brands()
			if a.So(tc.ExpectedErr(err), should.BeTrue) && err == nil {
				a.So(brands, should.Resemble, tc.ExpectedValue)
			}
		})
	}
}

func TestDeviceModels(t *testing.T) {
	for _, tc := range []struct {
		Name          string
		BrandID       string
		Fetcher       fetch.Interface
		ExpectedErr   func(err error) bool
		ExpectedValue interface{}
	}{
		{
			Name:    "Normal",
			BrandID: "thethingsproducts",
			Fetcher: validFetcher,
			ExpectedValue: map[string]ttnpb.EndDeviceModel{
				"thethingsuno": {
					BrandID: "thethingsproducts",
					ID:      "thethingsuno",
					Name:    "The Things Uno",
				},
			},
		},
		{
			Name:          "UnknownBrand",
			BrandID:       "unknown-brand",
			Fetcher:       validFetcher,
			ExpectedErr:   errors.IsNotFound,
			ExpectedValue: nil,
		},
		{
			Name:        "Invalid",
			BrandID:     "thethingsproducts",
			Fetcher:     invalidFetcher,
			ExpectedErr: errors.IsInvalidArgument,
		},
		{
			Name:        "Empty",
			BrandID:     "thethingsproducts",
			Fetcher:     emptyFetcher,
			ExpectedErr: errors.IsNotFound,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			repo := Client{Fetcher: tc.Fetcher}
			models, err := repo.DeviceModels(tc.BrandID)
			if tc.ExpectedErr == nil {
				if err != nil {
					t.Fatalf("Did not expect error but got %v", err)
				}
			} else if a.So(tc.ExpectedErr(err), should.BeTrue) && err == nil {
				a.So(models, should.Resemble, tc.ExpectedValue)
			}
		})
	}
}
func TestDeviceVersions(t *testing.T) {
	for _, tc := range []struct {
		Name          string
		BrandID       string
		ModelID       string
		Fetcher       fetch.Interface
		ExpectedErr   func(err error) bool
		ExpectedValue interface{}
	}{
		{
			Name:        "Normal",
			BrandID:     "thethingsproducts",
			ModelID:     "thethingsuno",
			Fetcher:     validFetcher,
			ExpectedErr: func(err error) bool { return err == nil },
			ExpectedValue: []ttnpb.EndDeviceVersion{
				{
					EndDeviceVersionIdentifiers: ttnpb.EndDeviceVersionIdentifiers{
						BrandID:         "thethingsproducts",
						ModelID:         "thethingsuno",
						HardwareVersion: "1.0",
						FirmwareVersion: "1.1",
					},
					Photos: []string{"front.jpg", "back.jpg"},
					DefaultFormatters: ttnpb.MessagePayloadFormatters{
						UpFormatter:            ttnpb.PayloadFormatter_FORMATTER_GRPC_SERVICE,
						UpFormatterParameter:   "hosted-service:1234",
						DownFormatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
						DownFormatterParameter: "function Encoder() { return { led: 1 } }",
					},
				},
			},
		},
		{
			Name:          "UnknownBrand",
			BrandID:       "unknown-brand",
			ModelID:       "unknown-model",
			Fetcher:       validFetcher,
			ExpectedErr:   errors.IsNotFound,
			ExpectedValue: nil,
		},
		{
			Name:          "UnknownModel",
			BrandID:       "thethingsproducts",
			ModelID:       "unknown-model",
			Fetcher:       validFetcher,
			ExpectedErr:   errors.IsNotFound,
			ExpectedValue: nil,
		},
		{
			Name:        "Invalid",
			BrandID:     "thethingsproducts",
			ModelID:     "thethingsuno",
			Fetcher:     invalidFetcher,
			ExpectedErr: errors.IsInvalidArgument,
		},
		{
			Name:        "Empty",
			BrandID:     "thethingsproducts",
			ModelID:     "thethingsuno",
			Fetcher:     emptyFetcher,
			ExpectedErr: errors.IsNotFound,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			repo := Client{Fetcher: tc.Fetcher}
			models, err := repo.DeviceVersions(tc.BrandID, tc.ModelID)
			if a.So(tc.ExpectedErr(err), should.BeTrue) && err == nil {
				a.So(models, should.Resemble, tc.ExpectedValue)
			}
		})
	}
}
