// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package lbscups

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

var errTest = errors.DefineInternal("test", "test error")

// testContext returns a context with authentication metadata for testing.
func testContext() context.Context {
	return metadata.NewIncomingContext(test.Context(), metadata.Pairs("authorization", "Bearer test-token"))
}

func TestCreateAPIKeys_Success(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	ids := &ttnpb.GatewayIdentifiers{
		Eui: types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0x01}.Bytes(),
	}

	callCount := 0
	mockAccess := &mockGatewayAccessClient{
		createAPIKeyFunc: func(_ context.Context, req *ttnpb.CreateGatewayAPIKeyRequest, _ ...grpc.CallOption) (*ttnpb.APIKey, error) {
			callCount++
			return &ttnpb.APIKey{
				Id:     fmt.Sprintf("key-%d", callCount),
				Key:    fmt.Sprintf("secret-%d", callCount),
				Name:   req.Name,
				Rights: req.Rights,
			}, nil
		},
	}

	upstream := &Upstream{
		component:     &mockComponent{allowInsecureFunc: func() bool { return true }},
		gatewayAccess: mockAccess,
	}

	cupsKey, lnsKey, err := upstream.createAPIKeys(testContext(), ids)

	a.So(err, should.BeNil)
	a.So(cupsKey, should.NotBeNil)
	a.So(cupsKey.Id, should.Equal, "key-1")
	a.So(cupsKey.Rights, should.Contain, ttnpb.Right_RIGHT_GATEWAY_INFO)
	a.So(cupsKey.Rights, should.Contain, ttnpb.Right_RIGHT_GATEWAY_SETTINGS_BASIC)
	a.So(cupsKey.Rights, should.Contain, ttnpb.Right_RIGHT_GATEWAY_READ_SECRETS)
	a.So(lnsKey, should.NotBeNil)
	a.So(lnsKey.Id, should.Equal, "key-2")
	a.So(lnsKey.Rights, should.Contain, ttnpb.Right_RIGHT_GATEWAY_LINK)
	a.So(len(lnsKey.Rights), should.Equal, 1)
}

func TestCreateAPIKeys_CUPSKeyCreationFails(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	ids := &ttnpb.GatewayIdentifiers{
		Eui: types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0x01}.Bytes(),
	}

	mockAccess := &mockGatewayAccessClient{
		createAPIKeyFunc: func(_ context.Context, _ *ttnpb.CreateGatewayAPIKeyRequest, _ ...grpc.CallOption) (*ttnpb.APIKey, error) {
			return nil, errTest.New()
		},
	}

	upstream := &Upstream{
		component:     &mockComponent{allowInsecureFunc: func() bool { return true }},
		gatewayAccess: mockAccess,
	}

	_, _, err := upstream.createAPIKeys(testContext(), ids)

	a.So(errors.IsAborted(err), should.BeTrue)
}

func TestCreateAPIKeys_LNSKeyCreationFails(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	ids := &ttnpb.GatewayIdentifiers{
		Eui: types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0x01}.Bytes(),
	}

	callCount := 0
	mockAccess := &mockGatewayAccessClient{
		createAPIKeyFunc: func(_ context.Context, req *ttnpb.CreateGatewayAPIKeyRequest, _ ...grpc.CallOption) (*ttnpb.APIKey, error) {
			callCount++
			if callCount == 1 {
				return &ttnpb.APIKey{
					Id:     "cups-key",
					Key:    "cups-secret",
					Name:   req.Name,
					Rights: req.Rights,
				}, nil
			}
			return nil, errTest.New()
		},
	}

	upstream := &Upstream{
		component:     &mockComponent{allowInsecureFunc: func() bool { return true }},
		gatewayAccess: mockAccess,
	}

	_, _, err := upstream.createAPIKeys(testContext(), ids)

	a.So(errors.IsAborted(err), should.BeTrue)
}

func TestCreateAPIKeys_VerifyRequestedRights(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	ids := &ttnpb.GatewayIdentifiers{
		Eui: types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0x01}.Bytes(),
	}

	var requests []*ttnpb.CreateGatewayAPIKeyRequest
	mockAccess := &mockGatewayAccessClient{
		createAPIKeyFunc: func(_ context.Context, req *ttnpb.CreateGatewayAPIKeyRequest, _ ...grpc.CallOption) (*ttnpb.APIKey, error) {
			requests = append(requests, req)
			return &ttnpb.APIKey{
				Id:     "key",
				Key:    "secret",
				Name:   req.Name,
				Rights: req.Rights,
			}, nil
		},
	}

	upstream := &Upstream{
		component:     &mockComponent{allowInsecureFunc: func() bool { return true }},
		gatewayAccess: mockAccess,
	}

	_, _, err := upstream.createAPIKeys(testContext(), ids)

	a.So(err, should.BeNil)
	a.So(len(requests), should.Equal, 2)

	// First request should be CUPS key
	cupsReq := requests[0]
	a.So(strings.HasPrefix(cupsReq.Name, "LBS CUPS Key"), should.BeTrue)
	a.So(cupsReq.Rights, should.Contain, ttnpb.Right_RIGHT_GATEWAY_INFO)
	a.So(cupsReq.Rights, should.Contain, ttnpb.Right_RIGHT_GATEWAY_SETTINGS_BASIC)
	a.So(cupsReq.Rights, should.Contain, ttnpb.Right_RIGHT_GATEWAY_READ_SECRETS)
	a.So(len(cupsReq.Rights), should.Equal, 3)

	// Second request should be LNS key
	lnsReq := requests[1]
	a.So(strings.HasPrefix(lnsReq.Name, "LBS LNS Key"), should.BeTrue)
	a.So(lnsReq.Rights, should.Contain, ttnpb.Right_RIGHT_GATEWAY_LINK)
	a.So(len(lnsReq.Rights), should.Equal, 1)
}

func TestDeleteAPIKeys_Success(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	ids := &ttnpb.GatewayIdentifiers{
		Eui: types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0x01}.Bytes(),
	}

	existingKeys := []*ttnpb.APIKey{
		{Id: "cups-key-1", Name: "LBS CUPS Key (TTGC claim), generated 2024-01-01"},
		{Id: "lns-key-1", Name: "LBS LNS Key (TTGC claim), generated 2024-01-01"},
		{Id: "other-key", Name: "Some other key"},
	}

	var deletedKeyIDs []string
	mockAccess := &mockGatewayAccessClient{
		listAPIKeysFunc: func(_ context.Context, _ *ttnpb.ListGatewayAPIKeysRequest, _ ...grpc.CallOption) (*ttnpb.APIKeys, error) {
			return &ttnpb.APIKeys{ApiKeys: existingKeys}, nil
		},
		deleteAPIKeyFunc: func(_ context.Context, req *ttnpb.DeleteGatewayAPIKeyRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
			deletedKeyIDs = append(deletedKeyIDs, req.KeyId)
			return &emptypb.Empty{}, nil
		},
	}

	upstream := &Upstream{
		component:     &mockComponent{allowInsecureFunc: func() bool { return true }},
		gatewayAccess: mockAccess,
	}

	err := upstream.deleteAPIKeys(testContext(), ids)

	a.So(err, should.BeNil)
	a.So(len(deletedKeyIDs), should.Equal, 2)
	a.So(deletedKeyIDs, should.Contain, "cups-key-1")
	a.So(deletedKeyIDs, should.Contain, "lns-key-1")
}

func TestDeleteAPIKeys_ListAPIKeysFails(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	ids := &ttnpb.GatewayIdentifiers{
		Eui: types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0x01}.Bytes(),
	}

	mockAccess := &mockGatewayAccessClient{
		listAPIKeysFunc: func(_ context.Context, _ *ttnpb.ListGatewayAPIKeysRequest, _ ...grpc.CallOption) (*ttnpb.APIKeys, error) {
			return nil, errTest.New()
		},
	}

	upstream := &Upstream{
		component:     &mockComponent{allowInsecureFunc: func() bool { return true }},
		gatewayAccess: mockAccess,
	}

	err := upstream.deleteAPIKeys(testContext(), ids)

	a.So(errors.IsAborted(err), should.BeTrue)
}

func TestDeleteAPIKeys_DeleteAPIKeyFailsContinues(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	ids := &ttnpb.GatewayIdentifiers{
		Eui: types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0x01}.Bytes(),
	}

	existingKeys := []*ttnpb.APIKey{
		{Id: "cups-key-1", Name: "LBS CUPS Key"},
		{Id: "lns-key-1", Name: "LBS LNS Key"},
	}

	var deletedKeyIDs []string
	callCount := 0
	mockAccess := &mockGatewayAccessClient{
		listAPIKeysFunc: func(_ context.Context, _ *ttnpb.ListGatewayAPIKeysRequest, _ ...grpc.CallOption) (*ttnpb.APIKeys, error) {
			return &ttnpb.APIKeys{ApiKeys: existingKeys}, nil
		},
		deleteAPIKeyFunc: func(_ context.Context, req *ttnpb.DeleteGatewayAPIKeyRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
			deletedKeyIDs = append(deletedKeyIDs, req.KeyId)
			callCount++
			if callCount == 1 {
				return nil, errTest.New()
			}
			return &emptypb.Empty{}, nil
		},
	}

	upstream := &Upstream{
		component:     &mockComponent{allowInsecureFunc: func() bool { return true }},
		gatewayAccess: mockAccess,
	}

	err := upstream.deleteAPIKeys(testContext(), ids)

	a.So(err, should.BeNil)
	// Both keys should be attempted to delete (even if first fails)
	a.So(len(deletedKeyIDs), should.Equal, 2)
}

func TestDeleteAPIKeys_SkipsEmptyNames(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	ids := &ttnpb.GatewayIdentifiers{
		Eui: types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0x01}.Bytes(),
	}

	existingKeys := []*ttnpb.APIKey{
		{Id: "empty-name-key", Name: ""},
		{Id: "cups-key-1", Name: "LBS CUPS Key"},
	}

	var deletedKeyIDs []string
	mockAccess := &mockGatewayAccessClient{
		listAPIKeysFunc: func(_ context.Context, _ *ttnpb.ListGatewayAPIKeysRequest, _ ...grpc.CallOption) (*ttnpb.APIKeys, error) {
			return &ttnpb.APIKeys{ApiKeys: existingKeys}, nil
		},
		deleteAPIKeyFunc: func(_ context.Context, req *ttnpb.DeleteGatewayAPIKeyRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
			deletedKeyIDs = append(deletedKeyIDs, req.KeyId)
			return &emptypb.Empty{}, nil
		},
	}

	upstream := &Upstream{
		component:     &mockComponent{allowInsecureFunc: func() bool { return true }},
		gatewayAccess: mockAccess,
	}

	err := upstream.deleteAPIKeys(testContext(), ids)

	a.So(err, should.BeNil)
	a.So(len(deletedKeyIDs), should.Equal, 1)
	a.So(deletedKeyIDs, should.Contain, "cups-key-1")
	a.So(deletedKeyIDs, should.NotContain, "empty-name-key")
}

func TestDeleteAPIKeys_SkipsNonLBSKeys(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	ids := &ttnpb.GatewayIdentifiers{
		Eui: types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0x01}.Bytes(),
	}

	existingKeys := []*ttnpb.APIKey{
		{Id: "other-key-1", Name: "Console generated key"},
		{Id: "other-key-2", Name: "CLI generated key"},
		{Id: "cups-key", Name: "LBS CUPS Key"},
	}

	var deletedKeyIDs []string
	mockAccess := &mockGatewayAccessClient{
		listAPIKeysFunc: func(_ context.Context, _ *ttnpb.ListGatewayAPIKeysRequest, _ ...grpc.CallOption) (*ttnpb.APIKeys, error) {
			return &ttnpb.APIKeys{ApiKeys: existingKeys}, nil
		},
		deleteAPIKeyFunc: func(_ context.Context, req *ttnpb.DeleteGatewayAPIKeyRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
			deletedKeyIDs = append(deletedKeyIDs, req.KeyId)
			return &emptypb.Empty{}, nil
		},
	}

	upstream := &Upstream{
		component:     &mockComponent{allowInsecureFunc: func() bool { return true }},
		gatewayAccess: mockAccess,
	}

	err := upstream.deleteAPIKeys(testContext(), ids)

	a.So(err, should.BeNil)
	a.So(len(deletedKeyIDs), should.Equal, 1)
	a.So(deletedKeyIDs, should.Contain, "cups-key")
}

// mockGatewayAccessClient implements ttnpb.GatewayAccessClient for testing.
type mockGatewayAccessClient struct {
	ttnpb.GatewayAccessClient

	createAPIKeyFunc func(
		ctx context.Context,
		req *ttnpb.CreateGatewayAPIKeyRequest,
		opts ...grpc.CallOption,
	) (*ttnpb.APIKey, error)
	listAPIKeysFunc func(
		ctx context.Context,
		req *ttnpb.ListGatewayAPIKeysRequest,
		opts ...grpc.CallOption,
	) (*ttnpb.APIKeys, error)
	deleteAPIKeyFunc func(
		ctx context.Context,
		req *ttnpb.DeleteGatewayAPIKeyRequest,
		opts ...grpc.CallOption,
	) (*emptypb.Empty, error)
}

func (m *mockGatewayAccessClient) CreateAPIKey(
	ctx context.Context,
	req *ttnpb.CreateGatewayAPIKeyRequest,
	opts ...grpc.CallOption,
) (*ttnpb.APIKey, error) {
	if m.createAPIKeyFunc != nil {
		return m.createAPIKeyFunc(ctx, req, opts...)
	}
	return nil, nil
}

func (m *mockGatewayAccessClient) ListAPIKeys(
	ctx context.Context,
	req *ttnpb.ListGatewayAPIKeysRequest,
	opts ...grpc.CallOption,
) (*ttnpb.APIKeys, error) {
	if m.listAPIKeysFunc != nil {
		return m.listAPIKeysFunc(ctx, req, opts...)
	}
	return nil, nil
}

func (m *mockGatewayAccessClient) DeleteAPIKey(
	ctx context.Context,
	req *ttnpb.DeleteGatewayAPIKeyRequest,
	opts ...grpc.CallOption,
) (*emptypb.Empty, error) {
	if m.deleteAPIKeyFunc != nil {
		return m.deleteAPIKeyFunc(ctx, req, opts...)
	}
	return nil, nil
}

// mockComponent implements the component interface for testing.
type mockComponent struct {
	allowInsecureFunc func() bool
}

func (m *mockComponent) GetTLSConfig(context.Context) tlsconfig.Config {
	return tlsconfig.Config{}
}

func (m *mockComponent) GetTLSClientConfig(context.Context, ...tlsconfig.Option) (*tls.Config, error) {
	return nil, nil
}

func (m *mockComponent) GetPeerConn(
	context.Context, ttnpb.ClusterRole, cluster.EntityIdentifiers,
) (*grpc.ClientConn, error) {
	return nil, nil
}

func (m *mockComponent) AllowInsecureForCredentials() bool {
	if m.allowInsecureFunc != nil {
		return m.allowInsecureFunc()
	}
	return true
}
