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

package ttjsv2_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	claimerrors "go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/enddevices/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/enddevices/ttjsv2"
	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/enddevices/ttjsv2/testdata"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var (
	serverAddress           = "127.0.0.1:0"
	client1ASID             = "client1.local"
	client2ASID             = "client2.local"
	claimAuthenticationCode = "SECRET"
	homeNSID                = types.EUI64{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}
	supportedJoinEUI        = types.EUI64{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C}
	supportedJoinEUIPrefix  = types.EUI64Prefix{
		EUI64:  supportedJoinEUI,
		Length: 64,
	}
	unsupportedJoinEUI = types.EUI64{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0D}
	unsupportedDevEUI  = types.EUI64{0x00, 0x04, 0xA3, 0x0B, 0x00, 0x1C, 0x05, 0xFF}
	devEUI             = types.EUI64{0x00, 0x04, 0xA3, 0x0B, 0x00, 0x1C, 0x05, 0x30}
	validEndDeviceIds  = &ttnpb.EndDeviceIdentifiers{
		DevEui:  devEUI.Bytes(),
		JoinEui: types.EUI64{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C}.Bytes(),
	}
	devEUI1 = types.EUI64{0x00, 0x04, 0xA3, 0x0B, 0x00, 0x1C, 0x05, 0x31}
	devEUI2 = types.EUI64{0x00, 0x04, 0xA3, 0x0B, 0x00, 0x1C, 0x05, 0x32}
	devEUI3 = types.EUI64{0x00, 0x04, 0xA3, 0x0B, 0x00, 0x1C, 0x05, 0x33}
)

func TestTTJS(t *testing.T) { //nolint:paralleltest
	a, ctx := test.New(t)
	lis, err := net.Listen("tcp", serverAddress)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	t.Cleanup(func() {
		lis.Close()
	})

	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(func() {
		cancel()
	})

	c := componenttest.NewComponent(t, &component.Config{})

	mockTTJS := mockTTJS{
		joinEUIPrefixes: []types.EUI64Prefix{
			supportedJoinEUIPrefix,
		},
		lis: lis,
		clients: map[string]clientData{
			client1ASID: {
				cert: testdata.Client1Cert,
			},
			client2ASID: {
				cert: testdata.Client2Cert,
			},
		},
		provisonedDevices: map[types.EUI64]device{
			devEUI: {
				claimAuthenticationCode: claimAuthenticationCode,
			},
		},
	}

	go mockTTJS.Start(ctx) //nolint:errcheck
	fetcher := fetch.FromFilesystem("testdata")

	client1 := ttjsv2.NewClient(c, fetcher, ttjsv2.Config{
		NetID: test.DefaultNetID,
		NSID:  &homeNSID,
		ASID:  client1ASID,
		JoinEUIPrefixes: []types.EUI64Prefix{
			supportedJoinEUIPrefix,
		},
		ConfigFile: ttjsv2.ConfigFile{
			URL: fmt.Sprintf("https://%s", lis.Addr().String()),
			TLS: ttjsv2.TLSConfig{
				RootCA:      "rootCA.pem",
				Source:      "file",
				Certificate: "clientcert-1.pem",
				Key:         "clientkey-1.pem",
			},
		},
	})

	// Check JoinEUI support.
	a.So(client1.SupportsJoinEUI(unsupportedJoinEUI), should.BeFalse)
	a.So(client1.SupportsJoinEUI(supportedJoinEUI), should.BeTrue)

	// Test Claiming
	for _, tc := range []struct { //nolint:paralleltest
		Name               string
		DevEUI             types.EUI64
		JoinEUI            types.EUI64
		AuthenticationCode string
		ErrorAssertion     func(err error) bool
	}{
		{
			Name:               "EmptyCAC",
			DevEUI:             devEUI,
			JoinEUI:            supportedJoinEUI,
			AuthenticationCode: "",
			ErrorAssertion: func(err error) bool {
				return errors.IsUnauthenticated(err)
			},
		},
		{
			Name:               "InvalidCAC",
			DevEUI:             devEUI,
			JoinEUI:            supportedJoinEUI,
			AuthenticationCode: "invalid",
			ErrorAssertion: func(err error) bool {
				return errors.IsUnauthenticated(err)
			},
		},
		{
			Name:               "NotProvisoned",
			DevEUI:             types.EUI64{},
			JoinEUI:            supportedJoinEUI,
			AuthenticationCode: claimAuthenticationCode,
			ErrorAssertion: func(err error) bool {
				return errors.IsNotFound(err)
			},
		},
		{
			Name:               "SuccessfulClaim",
			DevEUI:             devEUI,
			JoinEUI:            supportedJoinEUI,
			AuthenticationCode: claimAuthenticationCode,
		},
	} {
		t.Run(fmt.Sprintf("Claim/%s", tc.Name), func(t *testing.T) {
			err := client1.Claim(ctx, tc.JoinEUI, tc.DevEUI, tc.AuthenticationCode)
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(err), should.BeTrue)
			}
		})
	}

	// Claim locked.
	client2 := ttjsv2.NewClient(c, fetcher, ttjsv2.Config{
		NetID: test.DefaultNetID,
		NSID:  &homeNSID,
		ASID:  client2ASID,
		JoinEUIPrefixes: []types.EUI64Prefix{
			supportedJoinEUIPrefix,
		},
		ConfigFile: ttjsv2.ConfigFile{
			URL: fmt.Sprintf("https://%s", lis.Addr().String()),
			TLS: ttjsv2.TLSConfig{
				RootCA:      "rootCA.pem",
				Source:      "file",
				Certificate: "clientcert-2.pem",
				Key:         "clientkey-2.pem",
			},
		},
	})
	err = client2.Claim(ctx, supportedJoinEUI, devEUI, claimAuthenticationCode)
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	ret, err := client2.GetClaimStatus(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  devEUI.Bytes(),
		JoinEui: supportedJoinEUI.Bytes(),
	})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(ret, should.BeNil)
	err = client2.Unclaim(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  devEUI.Bytes(),
		JoinEui: supportedJoinEUI.Bytes(),
	})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)

	// Unclaim
	err = client1.Unclaim(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  devEUI.Bytes(),
		JoinEui: supportedJoinEUI.Bytes(),
	})
	a.So(err, should.BeNil)
	_, err = client2.GetClaimStatus(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  devEUI.Bytes(),
		JoinEui: supportedJoinEUI.Bytes(),
	})
	a.So(errors.IsNotFound(err), should.BeTrue)

	// Try to unclaim again
	err = client1.Unclaim(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  devEUI.Bytes(),
		JoinEui: supportedJoinEUI.Bytes(),
	})
	a.So(errors.IsNotFound(err), should.BeTrue)

	// Try to claim
	err = client1.Claim(ctx, supportedJoinEUI, devEUI, claimAuthenticationCode)
	a.So(err, should.BeNil)

	// Get valid status
	ret, err = client1.GetClaimStatus(ctx, validEndDeviceIds)
	a.So(err, should.BeNil)
	a.So(ret, should.NotBeNil)
	a.So(ret.EndDeviceIds, should.Resemble, validEndDeviceIds)
	a.So(ret.HomeNetId, should.Resemble, test.DefaultNetID.Bytes())
	var retHomeNSID types.EUI64
	err = retHomeNSID.Unmarshal(ret.HomeNsId)
	a.So(err, should.BeNil)
	a.So(retHomeNSID, should.Equal, homeNSID)
	a.So(err, should.BeNil)
	a.So(ret.VendorSpecific, should.BeNil)
}

func TestBatchOperations(t *testing.T) { // nolint:paralleltest
	a, ctx := test.New(t)
	lis, err := net.Listen("tcp", serverAddress)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	t.Cleanup(func() {
		lis.Close()
	})

	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(func() {
		cancel()
	})

	c := componenttest.NewComponent(t, &component.Config{})

	mockTTJS := mockTTJS{
		joinEUIPrefixes: []types.EUI64Prefix{
			supportedJoinEUIPrefix,
		},
		lis: lis,
		clients: map[string]clientData{
			client1ASID: {
				cert: testdata.Client1Cert,
			},
			client2ASID: {
				cert: testdata.Client2Cert,
			},
		},
		provisonedDevices: map[types.EUI64]device{
			devEUI1: {
				claimAuthenticationCode: claimAuthenticationCode,
			},
			devEUI2: {
				claimAuthenticationCode: claimAuthenticationCode,
			},
			devEUI3: {
				claimAuthenticationCode: claimAuthenticationCode,
			},
		},
	}

	go mockTTJS.Start(ctx) // nolint:errcheck
	fetcher := fetch.FromFilesystem("testdata")

	client := ttjsv2.NewClient(c, fetcher, ttjsv2.Config{
		NetID: test.DefaultNetID,
		NSID:  &homeNSID,
		ASID:  "localhost",
		JoinEUIPrefixes: []types.EUI64Prefix{
			supportedJoinEUIPrefix,
		},
		ConfigFile: ttjsv2.ConfigFile{
			URL: fmt.Sprintf("https://%s", lis.Addr().String()),
			TLS: ttjsv2.TLSConfig{
				RootCA:      "rootCA.pem",
				Source:      "file",
				Certificate: "clientcert-1.pem",
				Key:         "clientkey-1.pem",
			},
		},
	})

	client1 := ttjsv2.NewClient(c, fetcher, ttjsv2.Config{
		NetID: test.DefaultNetID,
		NSID:  &homeNSID,
		ASID:  "localhost1",
		JoinEUIPrefixes: []types.EUI64Prefix{
			supportedJoinEUIPrefix,
		},
		ConfigFile: ttjsv2.ConfigFile{
			URL: fmt.Sprintf("https://%s", lis.Addr().String()),
			TLS: ttjsv2.TLSConfig{
				RootCA:      "rootCA.pem",
				Source:      "file",
				Certificate: "clientcert-2.pem",
				Key:         "clientkey-2.pem",
			},
		},
	})

	// Check JoinEUI support.
	a.So(client.SupportsJoinEUI(unsupportedJoinEUI), should.BeFalse)
	a.So(client.SupportsJoinEUI(supportedJoinEUI), should.BeTrue)

	// Claim Devices.
	for _, dev := range []types.EUI64{devEUI1, devEUI2, devEUI3} {
		err = client.Claim(ctx, supportedJoinEUI, dev, claimAuthenticationCode)
		a.So(err, should.BeNil)
	}

	// Check Claim Status.
	for _, dev := range []types.EUI64{devEUI1, devEUI2, devEUI3} {
		ret, err := client.GetClaimStatus(ctx, &ttnpb.EndDeviceIdentifiers{
			DevEui:  dev.Bytes(),
			JoinEui: supportedJoinEUI.Bytes(),
		})
		a.So(err, should.BeNil)
		a.So(ret, should.NotBeNil)
		a.So(ret.EndDeviceIds, should.Resemble, &ttnpb.EndDeviceIdentifiers{
			DevEui:  dev.Bytes(),
			JoinEui: supportedJoinEUI.Bytes(),
		})
		a.So(ret.HomeNetId, should.Resemble, test.DefaultNetID.Bytes())
		var retHomeNSID types.EUI64
		err = retHomeNSID.Unmarshal(ret.HomeNsId)
		a.So(err, should.BeNil)
		a.So(retHomeNSID, should.Equal, homeNSID)
		a.So(ret.VendorSpecific, should.BeNil)
	}

	// Unclaim Devices.

	// Different client.
	err = client1.BatchUnclaim(ctx, []*ttnpb.EndDeviceIdentifiers{
		{
			DevEui:  devEUI1.Bytes(),
			JoinEui: supportedJoinEUI.Bytes(),
		},
		{
			DevEui:  devEUI2.Bytes(),
			JoinEui: supportedJoinEUI.Bytes(),
		},
		{
			DevEui:  devEUI3.Bytes(),
			JoinEui: supportedJoinEUI.Bytes(),
		},
	})
	a.So(err, should.NotBeNil)
	errs := claimerrors.DeviceErrors{}
	a.So(errors.As(err, &errs), should.BeTrue)
	a.So(errs.Errors, should.HaveLength, 3)

	// One Invalid device.
	devIds := &ttnpb.EndDeviceIdentifiers{
		DevEui:  unsupportedDevEUI.Bytes(),
		JoinEui: supportedJoinEUI.Bytes(),
	}
	err = client.BatchUnclaim(ctx, []*ttnpb.EndDeviceIdentifiers{
		{
			DevEui:  devEUI1.Bytes(),
			JoinEui: supportedJoinEUI.Bytes(),
		},
		devIds,
	})
	a.So(err, should.NotBeNil)
	a.So(errors.As(err, &errs), should.BeTrue)
	a.So(errs.Errors, should.HaveLength, 1)
	for eui := range errs.Errors {
		a.So(eui, should.Resemble, unsupportedDevEUI)
	}

	// Valid batch.
	err = client.BatchUnclaim(ctx, []*ttnpb.EndDeviceIdentifiers{
		{
			DevEui:  devEUI2.Bytes(),
			JoinEui: supportedJoinEUI.Bytes(),
		},
		{
			DevEui:  devEUI3.Bytes(),
			JoinEui: supportedJoinEUI.Bytes(),
		},
	})
	a.So(err, should.BeNil)
}
