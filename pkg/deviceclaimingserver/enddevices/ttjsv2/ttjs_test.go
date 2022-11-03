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

package ttjsv2

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var (
	serverAddress           = "127.0.0.1:0"
	password                = "secret"
	asID                    = "localhost"
	otherClientPassword     = "other-secret"
	otherClientASID         = "localhost-other"
	claimAuthenticationCode = "SECRET"
	homeNSID                = types.EUI64{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}
	supportedJoinEUI        = types.EUI64{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C}
	supportedJoinEUIPrefix  = types.EUI64Prefix{
		EUI64:  supportedJoinEUI,
		Length: 64,
	}
	unsupportedJoinEUI = types.EUI64{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0D}
	devEUI             = types.EUI64{0x00, 0x04, 0xA3, 0x0B, 0x00, 0x1C, 0x05, 0x30}
	validEndDeviceIds  = &ttnpb.EndDeviceIdentifiers{
		DevEui:  devEUI.Bytes(),
		JoinEui: types.EUI64{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C}.Bytes(),
	}
)

func TestTTJS(t *testing.T) { //nolint:tparallel
	t.Parallel()

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
			password: {
				asID: asID,
			},
			otherClientPassword: {
				asID: otherClientASID,
			},
		},
		provisonedDevices: map[types.EUI64]device{
			devEUI: {
				claimAuthenticationCode: claimAuthenticationCode,
			},
		},
	}

	go func() {
		if err := mockTTJS.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	// Valid Config
	ttJSConfig := Config{
		NetworkServer: NetworkServer{
			Hostname: "localhost",
			HomeNSID: &homeNSID,
		},
		NetID: test.DefaultNetID,
		JoinEUIPrefixes: []types.EUI64Prefix{
			supportedJoinEUIPrefix,
		},
		BasicAuth: BasicAuth{
			Username: asID,
			Password: "invalid",
		},
		URL: fmt.Sprintf("http://%s", lis.Addr().String()),
	}

	// Invalid client API key.
	unauthenticatedClient, err := ttJSConfig.NewClient(ctx, c)
	test.Must(unauthenticatedClient, err)
	err = unauthenticatedClient.Claim(ctx, supportedJoinEUI, devEUI, claimAuthenticationCode)
	a.So(errors.IsUnauthenticated(err), should.BeTrue)
	err = unauthenticatedClient.Unclaim(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  devEUI.Bytes(),
		JoinEui: supportedJoinEUI.Bytes(),
	})
	a.So(errors.IsUnauthenticated(err), should.BeTrue)
	ret, err := unauthenticatedClient.GetClaimStatus(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  devEUI.Bytes(),
		JoinEui: supportedJoinEUI.Bytes(),
	})
	a.So(errors.IsUnauthenticated(err), should.BeTrue)
	a.So(ret, should.BeNil)

	// With Valid Key
	ttJSConfig.Password = password
	client, err := ttJSConfig.NewClient(ctx, c)
	test.Must(client, err)

	// Check JoinEUI support.
	a.So(client.SupportsJoinEUI(unsupportedJoinEUI), should.BeFalse)
	a.So(client.SupportsJoinEUI(supportedJoinEUI), should.BeTrue)

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
			err := client.Claim(ctx, tc.JoinEUI, tc.DevEUI, tc.AuthenticationCode)
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
	otherClientConfig := Config{
		NetworkServer: NetworkServer{
			Hostname: "localhost",
			HomeNSID: &homeNSID,
		},
		NetID: test.DefaultNetID,
		JoinEUIPrefixes: []types.EUI64Prefix{
			supportedJoinEUIPrefix,
		},
		BasicAuth: BasicAuth{
			Username: otherClientASID,
			Password: otherClientPassword,
		},
		URL: fmt.Sprintf("http://%s", lis.Addr().String()),
	}
	otherClient, err := otherClientConfig.NewClient(ctx, c)
	test.Must(otherClient, err)
	err = otherClient.Claim(ctx, supportedJoinEUI, devEUI, claimAuthenticationCode)
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	ret, err = otherClient.GetClaimStatus(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  devEUI.Bytes(),
		JoinEui: supportedJoinEUI.Bytes(),
	})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(ret, should.BeNil)
	err = otherClient.Unclaim(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  devEUI.Bytes(),
		JoinEui: supportedJoinEUI.Bytes(),
	})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)

	// Unclaim
	err = client.Unclaim(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  devEUI.Bytes(),
		JoinEui: supportedJoinEUI.Bytes(),
	})
	a.So(err, should.BeNil)
	_, err = otherClient.GetClaimStatus(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  devEUI.Bytes(),
		JoinEui: supportedJoinEUI.Bytes(),
	})
	a.So(errors.IsNotFound(err), should.BeTrue)

	// Try to unclaim again
	err = client.Unclaim(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  devEUI.Bytes(),
		JoinEui: supportedJoinEUI.Bytes(),
	})
	a.So(errors.IsNotFound(err), should.BeTrue)

	// Try to claim
	err = client.Claim(ctx, supportedJoinEUI, devEUI, claimAuthenticationCode)
	a.So(err, should.BeNil)

	// Get valid status
	ret, err = client.GetClaimStatus(ctx, validEndDeviceIds)
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
