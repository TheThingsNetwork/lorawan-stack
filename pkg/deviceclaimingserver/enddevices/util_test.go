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

package enddevices

import (
	"context"
	"net/http"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/httpclient"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
	"google.golang.org/grpc"
)

var errMethodUnavailable = errors.DefineUnimplemented("method_unavailable", "method unavailable")

type mockEDCS struct {
	rights.EntityFetcher
	ttnpb.EndDeviceRegistryClient
}

func (mockEDCS) GetClusterID(context.Context) string {
	return "test"
}

func (mockEDCS) ClusterAddress(string) (string, error) {
	return "localhost", nil
}

func (mockEDCS) HTTPClient(ctx context.Context, opts ...httpclient.Option) (*http.Client, error) {
	return test.HTTPClientProvider.HTTPClient(ctx, opts...)
}

func (mockEDCS) GetBaseConfig(context.Context) config.ServiceBase {
	return config.ServiceBase{}
}

func (mockEDCS) GetPeerConn(context.Context, ttnpb.ClusterRole, cluster.EntityIdentifiers) (*grpc.ClientConn, error) {
	return nil, errMethodUnavailable.New()
}

func (mockEDCS) AuthInfo(context.Context) (*ttnpb.AuthInfoResponse, error) {
	return nil, errMethodUnavailable.New()
}

func (mockEDCS) AllowInsecureForCredentials() bool {
	return true
}

// Get implements EndDeviceRegistryClient.
func (mockEDCS) Get(_ context.Context, in *ttnpb.GetEndDeviceRequest, _ ...grpc.CallOption) (*ttnpb.EndDevice, error) {
	return &ttnpb.EndDevice{
		Ids: in.EndDeviceIds,
	}, nil
}

// ApplicationRights implements the Fetcher interface.
func (mockEDCS) ApplicationRights(context.Context, *ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	return &ttnpb.Rights{
		Rights: []ttnpb.Right{
			ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
		},
	}, nil
}

// SupportsJoinEUI implements EndDeviceClaimingServer.
func (mockEDCS) SupportsJoinEUI(types.EUI64) bool {
	return false
}

// RegisterRoutes implements EndDeviceClaimingServer.
func (mockEDCS) RegisterRoutes(*web.Server) {
}

// Claim implements EndDeviceClaimingServer.
func (mockEDCS) Claim(context.Context, *ttnpb.ClaimEndDeviceRequest) (ids *ttnpb.EndDeviceIdentifiers, err error) {
	return nil, errMethodUnavailable.New()
}

// Unclaim implements EndDeviceClaimingServer.
func (mockEDCS) Unclaim(context.Context, *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	return nil, errMethodUnavailable.New()
}

// GetInfoByJoinEUI implements EndDeviceClaimingServer.
func (mockEDCS) GetInfoByJoinEUI(
	context.Context,
	*ttnpb.GetInfoByJoinEUIRequest,
) (*ttnpb.GetInfoByJoinEUIResponse, error) {
	return nil, errMethodUnavailable.New()
}

// GetClaimStatus implements EndDeviceClaimingServer.
func (mockEDCS) GetClaimStatus(context.Context, *ttnpb.EndDeviceIdentifiers) (*ttnpb.GetClaimStatusResponse, error) {
	return nil, errMethodUnavailable.New()
}
