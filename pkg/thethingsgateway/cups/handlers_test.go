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

package cups

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	. "go.thethings.network/lorawan-stack/pkg/component/test"
)

const (
	testFirmwarePath  = "https://thethingsproducts.blob.core.windows.net/the-things-gateway/v1"
	testUpdateChannel = "stable"
)

func TestAdaptUpdateChannel(t *testing.T) {
	var conf Config
	conf.Default.UpdateChannel = testUpdateChannel
	conf.Default.FirmwareURL = testFirmwarePath
	s, err := conf.NewServer(NewComponent(t, &component.Config{}))
	if err != nil {
		t.Error(err)
	}

	for _, tt := range []struct {
		Name            string
		Channel         string
		ExpectedChannel string
	}{
		{
			Name:            "Empty channel",
			Channel:         "",
			ExpectedChannel: fmt.Sprintf("%v/%v", testFirmwarePath, "stable"),
		},
		{
			Name:            "Default stable channel",
			Channel:         "stable",
			ExpectedChannel: fmt.Sprintf("%v/%v", testFirmwarePath, "stable"),
		},
		{
			Name:            "Default beta channel",
			Channel:         "beta",
			ExpectedChannel: fmt.Sprintf("%v/%v", testFirmwarePath, "beta"),
		},
		{
			Name:            "Custom update channel",
			Channel:         "http://example.com/fake-firmware",
			ExpectedChannel: "http://example.com/fake-firmware",
		},
		{
			Name:            "Channel URL misspell",
			Channel:         "htp://example.com/stable",
			ExpectedChannel: "htp://example.com/stable",
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			a := assertions.New(t)
			a.So(s.adaptUpdateChannel(tt.Channel), assertions.ShouldEqual, tt.ExpectedChannel)
		})
	}
}

func TestAdaptGatewayAddress(t *testing.T) {
	for _, tt := range []struct {
		Name    string
		Address string
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
				a.So(address, assertions.ShouldEqual, "mqtts://localhost:8882")
			},
		},
		{
			Name:    "Host and port, no scheme",
			Address: "localhost:8881",
			Assert: func(a *assertions.Assertion, address string, err error) {
				a.So(err, assertions.ShouldBeNil)
				a.So(address, assertions.ShouldEqual, "mqtts://localhost:8881")
			},
		},
		{
			Name:    "Full mqtts address",
			Address: "mqtts://localhost:8882",
			Assert: func(a *assertions.Assertion, address string, err error) {
				a.So(err, assertions.ShouldBeNil)
				a.So(address, assertions.ShouldEqual, "mqtts://localhost:8882")
			},
		},
		{
			Name:    "Full mqtt address",
			Address: "mqtt://localhost:1882",
			Assert: func(a *assertions.Assertion, address string, err error) {
				a.So(err, assertions.ShouldBeNil)
				a.So(address, assertions.ShouldEqual, "mqtt://localhost:1882")
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
			address, err := adaptGatewayAddress(tt.Address)
			tt.Assert(a, address, err)
		})
	}
}

func TestAdaptAuthorization(t *testing.T) {
	for _, tt := range []struct {
		Name          string
		Authorization string
		Assert        func(*assertions.Assertion, string)
	}{
		{
			Name:          "Empty Authorization",
			Authorization: "",
			Assert: func(a *assertions.Assertion, auth string) {
				a.So(auth, assertions.ShouldEqual, "")
			},
		},
		{
			Name:          "Key formatted authorization",
			Authorization: "Key asd",
			Assert: func(a *assertions.Assertion, auth string) {
				a.So(auth, assertions.ShouldEqual, "asd")
			},
		},
		{
			Name:          "Bearer formatted authorization",
			Authorization: "Bearer efg",
			Assert: func(a *assertions.Assertion, auth string) {
				a.So(auth, assertions.ShouldEqual, "efg")
			},
		},
		{
			Name:          "Direct API Key",
			Authorization: "InvalidKeyFormat",
			Assert: func(a *assertions.Assertion, auth string) {
				a.So(auth, assertions.ShouldEqual, "InvalidKeyFormat")
			},
		},
		{
			Name:          "Esoteric authorization",
			Authorization: "APIKey asd",
			Assert: func(a *assertions.Assertion, auth string) {
				a.So(auth, assertions.ShouldEqual, "asd")
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			a := assertions.New(t)
			auth := adaptAuthorization(tt.Authorization)
			tt.Assert(a, auth)
		})
	}
}
