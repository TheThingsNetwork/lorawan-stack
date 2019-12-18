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

package gcsv2

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/pkg/component/test"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestSetTTKGFirmwareURL(t *testing.T) {
	buildConfig := func(firmwareURL, channel string) TheThingsGatewayConfig {
		var c TheThingsGatewayConfig
		c.Default.FirmwareURL = firmwareURL
		c.Default.UpdateChannel = channel
		return c
	}

	for _, tt := range []struct {
		Name          string
		Config        TheThingsGatewayConfig
		UpdateChannel string
		ExpectedURL   string
	}{
		{
			Name:          "No config, no channel",
			UpdateChannel: "",
			ExpectedURL:   "https://thethingsproducts.blob.core.windows.net/the-things-gateway/v1/stable",
		},
		{
			Name:          "No config, beta channel",
			UpdateChannel: "beta",
			ExpectedURL:   "https://thethingsproducts.blob.core.windows.net/the-things-gateway/v1/beta",
		},
		{
			Name:          "Config with defaults, no channel",
			Config:        buildConfig("https://firmware.net/the-things-gateway/v1", "channel"),
			UpdateChannel: "",
			ExpectedURL:   "https://firmware.net/the-things-gateway/v1/channel",
		},
		{
			Name:          "Config with defaults, beta channel",
			Config:        buildConfig("https://firmware.net/the-things-gateway/v1", "channel"),
			UpdateChannel: "beta",
			ExpectedURL:   "https://firmware.net/the-things-gateway/v1/beta",
		},
		{
			Name:          "No config, full URL",
			UpdateChannel: "https://firmware.net/the-things-gateway/v1/channel",
			ExpectedURL:   "https://firmware.net/the-things-gateway/v1/channel",
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			a := assertions.New(t)
			s := New(componenttest.NewComponent(t, &component.Config{}), WithTheThingsGatewayConfig(tt.Config))
			var res gatewayInfoResponse
			s.setTTKGFirmwareURL(&res, &ttnpb.Gateway{UpdateChannel: tt.UpdateChannel})
			a.So(res.FirmwareURL, should.Equal, tt.ExpectedURL)
		})
	}
}

func TestInferMQTTAddress(t *testing.T) {
	for _, tt := range []struct {
		Name    string
		Address string
		Config  TheThingsGatewayConfig
		Assert  func(*assertions.Assertion, string, error)
	}{
		{
			Name:    "Empty address",
			Address: "",
			Assert: func(a *assertions.Assertion, address string, err error) {
				a.So(err, assertions.ShouldBeNil)
				a.So(address, assertions.ShouldEqual, "")
			},
		},
		{
			Name:    "Only host, no port or scheme",
			Address: "localhost",
			Assert: func(a *assertions.Assertion, address string, err error) {
				a.So(err, assertions.ShouldBeNil)
				a.So(address, assertions.ShouldEqual, "mqtts://localhost:8881")
			},
		},
		{
			Name:    "Host and MQTT port, no scheme",
			Address: "localhost:1881",
			Assert: func(a *assertions.Assertion, address string, err error) {
				a.So(err, assertions.ShouldBeNil)
				a.So(address, assertions.ShouldEqual, "mqtt://localhost:1881")
			},
		},
		{
			Name:    "Host and MQTTS port, no scheme",
			Address: "localhost:8881",
			Assert: func(a *assertions.Assertion, address string, err error) {
				a.So(err, assertions.ShouldBeNil)
				a.So(address, assertions.ShouldEqual, "mqtts://localhost:8881")
			},
		},
		{
			Name:    "Full mqtts address",
			Address: "mqtts://localhost:8871",
			Assert: func(a *assertions.Assertion, address string, err error) {
				a.So(err, assertions.ShouldBeNil)
				a.So(address, assertions.ShouldEqual, "mqtts://localhost:8871")
			},
		},
		{
			Name:    "Full mqtt address",
			Address: "mqtt://localhost:1871",
			Assert: func(a *assertions.Assertion, address string, err error) {
				a.So(err, assertions.ShouldBeNil)
				a.So(address, assertions.ShouldEqual, "mqtt://localhost:1871")
			},
		},
		{
			Name:    "Full http address",
			Address: "http://localhost",
			Assert: func(a *assertions.Assertion, address string, err error) {
				a.So(err, assertions.ShouldBeNil)
				a.So(address, assertions.ShouldEqual, "http://localhost")
			},
		},
		{
			Name:    "Invalid port format",
			Address: "localhost::zzz",
			Assert: func(a *assertions.Assertion, address string, err error) {
				a.So(err, assertions.ShouldNotBeNil)
				a.So(address, assertions.ShouldEqual, "")
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			a := assertions.New(t)
			s := New(componenttest.NewComponent(t, &component.Config{}), WithTheThingsGatewayConfig(tt.Config))
			address, err := s.inferMQTTAddress(tt.Address)
			tt.Assert(a, address, err)
		})
	}
}
