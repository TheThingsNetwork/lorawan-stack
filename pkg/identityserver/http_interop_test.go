// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package identityserver

import (
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/interop"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestInteropServer(t *testing.T) {
	t.Parallel()

	p := &storetest.Population{}

	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, _ *grpc.ClientConn) {
		srv := &interopServer{
			IdentityServer: is,
		}

		joinEUI := types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}
		devEUI := types.EUI64{8, 7, 6, 5, 4, 3, 2, 1}
		id := &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{
				ApplicationId: "test-app-id",
			},
			DeviceId: "test-device-id",
			JoinEui:  joinEUI.Bytes(),
			DevEui:   devEUI.Bytes(),
		}

		_, err := is.store.CreateEndDevice(ctx, &ttnpb.EndDevice{
			Ids:                  id,
			NetworkServerAddress: "thethings.example.com",
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		defer is.store.DeleteEndDevice(ctx, id)

		ctx := interop.NewContextWithNetworkServerAuthInfo(ctx, &interop.NetworkServerAuthInfo{
			NetID:     types.NetID{0x0, 0x0, 0x0},
			Addresses: []string{"localhost"},
		})

		// Test a known device with Backend Interfaces 1.0
		ans, err := srv.HomeNSRequest(ctx, &interop.HomeNSReq{
			NsJsMessageHeader: interop.NsJsMessageHeader{
				MessageHeader: interop.MessageHeader{
					ProtocolVersion: interop.ProtocolV1_0,
					TransactionID:   42,
					MessageType:     interop.MessageTypeHomeNSReq,
				},
				SenderID:   interop.NetID{0x0, 0x0, 0x0},
				ReceiverID: interop.EUI64(joinEUI),
			},
			DevEUI: interop.EUI64(devEUI),
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		a.So(ans.HNetID, should.Equal, interop.NetID(test.DefaultNetID))
		a.So(ans.HNSID, should.BeNil) // Backend Interfaces 1.0 does not support this

		// Test a known device with Backend Interfaces 1.1
		ans, err = srv.HomeNSRequest(ctx, &interop.HomeNSReq{
			NsJsMessageHeader: interop.NsJsMessageHeader{
				MessageHeader: interop.MessageHeader{
					ProtocolVersion: interop.ProtocolV1_1,
					TransactionID:   42,
					MessageType:     interop.MessageTypeHomeNSReq,
				},
				SenderID:   interop.NetID{0x0, 0x0, 0x0},
				ReceiverID: interop.EUI64(joinEUI),
			},
			DevEUI: interop.EUI64(devEUI),
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		a.So(ans.HNetID, should.Equal, interop.NetID(test.DefaultNetID))
		a.So(ans.HNSID, should.Resemble, (*interop.EUI64)(&test.DefaultNSID))

		// Test an unknown device
		_, err = srv.HomeNSRequest(ctx, &interop.HomeNSReq{
			NsJsMessageHeader: interop.NsJsMessageHeader{
				MessageHeader: interop.MessageHeader{
					ProtocolVersion: interop.ProtocolV1_1,
					TransactionID:   42,
					MessageType:     interop.MessageTypeHomeNSReq,
				},
				SenderID:   interop.NetID{0x0, 0x0, 0x0},
				ReceiverID: interop.EUI64(joinEUI),
			},
			DevEUI: interop.EUI64{8, 8, 8, 8, 8, 8, 8, 8},
		})
		if !a.So(errors.IsNotFound(err), should.BeTrue) {
			t.FailNow()
		}
	}, withPrivateTestDatabase(p))
}
