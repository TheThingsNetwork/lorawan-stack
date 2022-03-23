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

package ttjs

import (
	"context"
	"fmt"
	"net"
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
	apiVersion              = "v1"
	asID                    = "localhost"
	otherClientPassword     = "other-secret"
	otherClientASID         = "localhost-other"
	claimAuthenticationCode = "SECRET"
	nsAddress               = "localhost"
	homeNSID                = "1122334455667788"
	supportedJoinEUI        = types.EUI64{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C}
	supportedJoinEUIPrefix  = types.EUI64Prefix{
		EUI64:  supportedJoinEUI,
		Length: 64,
	}
	unsupportedJoinEUI       = types.EUI64{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0D}
	unsupportedJoinEUIPrefix = types.EUI64Prefix{
		EUI64:  unsupportedJoinEUI,
		Length: 64,
	}
	devEUI            = types.EUI64{0x00, 0x04, 0xA3, 0x0B, 0x00, 0x1C, 0x05, 0x30}
	validEndDeviceIds = &ttnpb.EndDeviceIdentifiers{
		DevEui:  &devEUI,
		JoinEui: &types.EUI64{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C},
	}
	tenantID     = "test-os"
	targetAppIDs = &ttnpb.ApplicationIdentifiers{
		ApplicationId: "test-app",
	}
	targetDeviceID = "test-dev"
)

func TestTTJS(t *testing.T) {
	a, ctx := test.New(t)
	lis, err := net.Listen("tcp", serverAddress)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer lis.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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

	go func() error {
		return mockTTJS.Start(ctx, apiVersion)
	}()

	// Invalid Config
	invalidConfig := Config{
		HomeNSIDs: map[string]string{
			"localhost": "1234",
		},
		TenantID: tenantID,
		NetID:    test.DefaultNetID,
		JoinEUIPrefixes: []types.EUI64Prefix{
			supportedJoinEUIPrefix,
		},
		ClaimingAPIVersion: apiVersion,
		BasicAuth: BasicAuth{
			Username: asID,
			Password: "invalid",
		},
		URL: fmt.Sprintf("http://%s", lis.Addr().String()),
	}

	cl, err := invalidConfig.NewClient(ctx, c)
	a.So(errors.IsInvalidArgument(err), should.BeTrue)
	a.So(cl, should.BeNil)

	// Valid Config
	ttJSConfig := Config{
		HomeNSIDs: map[string]string{
			"localhost": homeNSID,
		},
		TenantID: tenantID,
		NetID:    test.DefaultNetID,
		JoinEUIPrefixes: []types.EUI64Prefix{
			supportedJoinEUIPrefix,
		},
		ClaimingAPIVersion: apiVersion,
		BasicAuth: BasicAuth{
			Username: asID,
			Password: "invalid",
		},
		URL: fmt.Sprintf("http://%s", lis.Addr().String()),
	}

	// Invalid client API key.
	unauthenticatedClient, err := ttJSConfig.NewClient(ctx, c)
	test.Must(unauthenticatedClient, err)
	_, err = unauthenticatedClient.Claim(ctx, &ttnpb.ClaimEndDeviceRequest{
		SourceDevice: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers_{
			AuthenticatedIdentifiers: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers{
				JoinEui:            supportedJoinEUI,
				DevEui:             devEUI,
				AuthenticationCode: claimAuthenticationCode,
			},
		},
		TargetNetworkServerAddress: nsAddress,
		TargetApplicationIds:       targetAppIDs,
		TargetDeviceId:             targetDeviceID,
	})
	a.So(errors.IsUnauthenticated(err), should.BeTrue)
	err = unauthenticatedClient.Unclaim(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  &devEUI,
		JoinEui: &supportedJoinEUI,
	})
	a.So(errors.IsUnauthenticated(err), should.BeTrue)
	ret, err := unauthenticatedClient.GetClaimStatus(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  &devEUI,
		JoinEui: &supportedJoinEUI,
	})
	a.So(errors.IsUnauthenticated(err), should.BeTrue)
	a.So(ret, should.BeNil)

	// With Valid Key
	ttJSConfig.Password = password
	client, err := ttJSConfig.NewClient(ctx, c, WithQRGeneratorClient(mockQRGClient{}))
	test.Must(client, err)

	// Check JoinEUI support.
	a.So(client.SupportsJoinEUI(unsupportedJoinEUI), should.BeFalse)
	a.So(client.SupportsJoinEUI(supportedJoinEUI), should.BeTrue)

	// Test Claiming
	for _, tc := range []struct {
		Name               string
		Req                *ttnpb.ClaimEndDeviceRequest
		AuthenticationCode []byte
		ErrorAssertion     func(err error) bool
	}{
		{
			Name: "EmptyCAC",
			Req: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers{
						JoinEui:            supportedJoinEUI,
						DevEui:             devEUI,
						AuthenticationCode: "invalid",
					},
				},
				TargetNetworkServerAddress: nsAddress,
				TargetApplicationIds:       targetAppIDs,
				TargetDeviceId:             targetDeviceID,
			},
			ErrorAssertion: func(err error) bool {
				return errors.IsUnauthenticated(err)
			},
		},
		{
			Name: "InvalidCAC",
			Req: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers{
						JoinEui:            supportedJoinEUI,
						DevEui:             devEUI,
						AuthenticationCode: "invalid",
					},
				},
				TargetNetworkServerAddress: nsAddress,
				TargetApplicationIds:       targetAppIDs,
				TargetDeviceId:             targetDeviceID,
			},
			ErrorAssertion: func(err error) bool {
				return errors.IsUnauthenticated(err)
			},
		},
		{
			Name: "NoTargetNSID",
			Req: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers{
						JoinEui:            supportedJoinEUI,
						DevEui:             devEUI,
						AuthenticationCode: claimAuthenticationCode,
					},
				},
				TargetApplicationIds: targetAppIDs,
				TargetDeviceId:       targetDeviceID,
			},
			ErrorAssertion: func(err error) bool {
				return errors.IsInvalidArgument(err)
			},
		},
		{
			Name: "NotProvisoned",
			Req: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers{
						JoinEui:            supportedJoinEUI,
						DevEui:             types.EUI64{},
						AuthenticationCode: claimAuthenticationCode,
					},
				},
				TargetNetworkServerAddress: nsAddress,
				TargetApplicationIds:       targetAppIDs,
				TargetDeviceId:             targetDeviceID,
			},
			ErrorAssertion: func(err error) bool {
				return errors.IsNotFound(err)
			},
		},
		{
			Name: "SuccessfulClaim",
			Req: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers{
						JoinEui:            supportedJoinEUI,
						DevEui:             devEUI,
						AuthenticationCode: claimAuthenticationCode,
					},
				},
				TargetNetworkServerAddress: nsAddress,
				TargetApplicationIds:       targetAppIDs,
				TargetDeviceId:             targetDeviceID,
			},
		},
		{
			Name: "SuccessfulClaimWithQRCode",
			Req: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_QrCode{
					QrCode: []byte("LW:D0:800000000000000C:0004A30B001C0530:42FFFF42:OSECRET"),
				},
				TargetNetworkServerAddress: nsAddress,
				TargetApplicationIds:       targetAppIDs,
				TargetDeviceId:             targetDeviceID,
			},
		},
	} {
		t.Run(fmt.Sprintf("Claim/%s", tc.Name), func(t *testing.T) {
			_, err := client.Claim(ctx, tc.Req)
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
		HomeNSIDs: map[string]string{
			"localhost": homeNSID,
		},
		NetID: test.DefaultNetID,
		JoinEUIPrefixes: []types.EUI64Prefix{
			supportedJoinEUIPrefix,
		},
		ClaimingAPIVersion: apiVersion,
		BasicAuth: BasicAuth{
			Username: otherClientASID,
			Password: otherClientPassword,
		},
		URL: fmt.Sprintf("http://%s", lis.Addr().String()),
	}
	otherClient, err := otherClientConfig.NewClient(ctx, c)
	test.Must(otherClient, err)
	_, err = otherClient.Claim(ctx, &ttnpb.ClaimEndDeviceRequest{
		SourceDevice: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers_{
			AuthenticatedIdentifiers: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers{
				JoinEui:            supportedJoinEUI,
				DevEui:             devEUI,
				AuthenticationCode: claimAuthenticationCode,
			},
		},
		TargetNetworkServerAddress: nsAddress,
		TargetApplicationIds:       targetAppIDs,
		TargetDeviceId:             targetDeviceID,
	})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	ret, err = otherClient.GetClaimStatus(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  &devEUI,
		JoinEui: &supportedJoinEUI,
	})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(ret, should.BeNil)
	err = otherClient.Unclaim(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  &devEUI,
		JoinEui: &supportedJoinEUI,
	})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)

	// Unclaim
	err = client.Unclaim(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  &devEUI,
		JoinEui: &supportedJoinEUI,
	})
	a.So(err, should.BeNil)
	ret, err = otherClient.GetClaimStatus(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  &devEUI,
		JoinEui: &supportedJoinEUI,
	})
	a.So(errors.IsNotFound(err), should.BeTrue)

	// Try to unclaim again
	err = client.Unclaim(ctx, &ttnpb.EndDeviceIdentifiers{
		DevEui:  &devEUI,
		JoinEui: &supportedJoinEUI,
	})
	a.So(errors.IsNotFound(err), should.BeTrue)

	// Try to claim
	ids, err := client.Claim(ctx, &ttnpb.ClaimEndDeviceRequest{
		SourceDevice: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers_{
			AuthenticatedIdentifiers: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers{
				JoinEui:            supportedJoinEUI,
				DevEui:             devEUI,
				AuthenticationCode: claimAuthenticationCode,
			},
		},
		TargetNetworkServerAddress: nsAddress,
		TargetApplicationIds:       targetAppIDs,
		TargetDeviceId:             targetDeviceID,
	})
	a.So(err, should.BeNil)
	a.So(ids, should.NotBeNil)

	// Get valid status
	ret, err = client.GetClaimStatus(ctx, validEndDeviceIds)
	a.So(err, should.BeNil)
	a.So(ret, should.NotBeNil)
	a.So(ret.EndDeviceIds, should.Resemble, validEndDeviceIds)
	a.So(*ret.HomeNetId, should.Resemble, test.DefaultNetID)
	a.So(ret.HomeNsId.String(), should.Equal, homeNSID)
	a.So(err, should.BeNil)
	a.So(ret.VendorSpecific, should.NotBeNil)
	a.So(ret.VendorSpecific.OrganizationUniqueIdentifier, should.Equal, 0xec656e)
}
