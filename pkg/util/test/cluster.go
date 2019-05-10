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

package test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// MockPeer is a mock cluster.Peer used for testing.
type MockPeer struct {
	NameFunc    func() string
	ConnFunc    func() *grpc.ClientConn
	HasRoleFunc func(ttnpb.PeerInfo_Role) bool
	RolesFunc   func() []ttnpb.PeerInfo_Role
	TagsFunc    func() map[string]string
}

// Name calls NameFunc if set and returns empty string otherwise.
func (m MockPeer) Name() string {
	if m.NameFunc == nil {
		return ""
	}
	return m.NameFunc()
}

// Conn calls ConnFunc if set and returns nil otherwise.
func (m MockPeer) Conn() *grpc.ClientConn {
	if m.ConnFunc == nil {
		return nil
	}
	return m.ConnFunc()
}

// HasRole calls HasRoleFunc if set and returns false otherwise.
func (m MockPeer) HasRole(r ttnpb.PeerInfo_Role) bool {
	if m.HasRoleFunc == nil {
		return false
	}
	return m.HasRoleFunc(r)
}

// Roles calls RolesFunc if set and returns nil otherwise.
func (m MockPeer) Roles() []ttnpb.PeerInfo_Role {
	if m.RolesFunc == nil {
		return nil
	}
	return m.RolesFunc()
}

// Tags calls TagsFunc if set and returns nil otherwise.
func (m MockPeer) Tags() map[string]string {
	if m.TagsFunc == nil {
		return nil
	}
	return m.TagsFunc()
}

// NewGRPCServerPeer creates a new MockPeer with ConnFunc, which always returns the same loopback connection to the server itself.
// srv is the implementation of the gRPC interface.
// registrators represents a slice of functions, which register the gRPC interface implementation at a gRPC server.
func NewGRPCServerPeer(ctx context.Context, srv interface{}, registrators ...interface{}) (*MockPeer, error) {
	grpcSrv := grpc.NewServer()
	for _, r := range registrators {
		reflect.ValueOf(r).Call([]reflect.Value{
			reflect.ValueOf(grpcSrv),
			reflect.ValueOf(srv),
		})
	}
	conn, err := rpcserver.StartLoopback(ctx, grpcSrv)
	if err != nil {
		return nil, err
	}
	return &MockPeer{
		ConnFunc: func() *grpc.ClientConn { return conn },
	}, nil
}

// MockCluster is a mock cluster.Cluster used for testing.
type MockCluster struct {
	JoinFunc               func() error
	LeaveFunc              func() error
	GetPeersFunc           func(ctx context.Context, role ttnpb.PeerInfo_Role) []cluster.Peer
	GetPeerFunc            func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer
	ClaimIDsFunc           func(ctx context.Context, ids ttnpb.Identifiers) error
	UnclaimIDsFunc         func(ctx context.Context, ids ttnpb.Identifiers) error
	TLSFunc                func() bool
	AuthFunc               func() grpc.CallOption
	WithVerifiedSourceFunc func(ctx context.Context) context.Context
}

// Join calls JoinFunc if set and returns nil otherwise.
func (m MockCluster) Join() error {
	if m.JoinFunc == nil {
		return nil
	}
	return m.JoinFunc()
}

// Leave calls LeaveFunc if set and returns nil otherwise.
func (m MockCluster) Leave() error {
	if m.LeaveFunc == nil {
		return nil
	}
	return m.LeaveFunc()
}

// GetPeers calls GetPeersFunc if set and returns nil otherwise.
func (m MockCluster) GetPeers(ctx context.Context, role ttnpb.PeerInfo_Role) []cluster.Peer {
	if m.GetPeersFunc == nil {
		return nil
	}
	return m.GetPeersFunc(ctx, role)
}

// GetPeer calls GetPeerFunc if set and returns nil otherwise.
func (m MockCluster) GetPeer(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer {
	if m.GetPeerFunc == nil {
		return nil
	}
	return m.GetPeerFunc(ctx, role, ids)
}

// ClaimIDs calls ClaimIDsFunc if set and returns nil otherwise.
func (m MockCluster) ClaimIDs(ctx context.Context, ids ttnpb.Identifiers) error {
	if m.ClaimIDsFunc == nil {
		return nil
	}
	return m.ClaimIDsFunc(ctx, ids)
}

// UnclaimIDs calls UnclaimIDsFunc if set and returns nil otherwise.
func (m MockCluster) UnclaimIDs(ctx context.Context, ids ttnpb.Identifiers) error {
	if m.UnclaimIDsFunc == nil {
		return nil
	}
	return m.UnclaimIDsFunc(ctx, ids)
}

// TLS calls TLSFunc if s et and returns false otherwise.
func (m MockCluster) TLS() bool {
	if m.TLSFunc == nil {
		return false
	}
	return m.TLSFunc()
}

// Auth calls AuthFunc if set and returns grpc.EmptyCallOption{} otherwise.
func (m MockCluster) Auth() grpc.CallOption {
	if m.AuthFunc == nil {
		return grpc.EmptyCallOption{}
	}
	return m.AuthFunc()
}

// WithVerifiedSource calls WithVerifiedSourceFunc if set and returns ctx otherwise.
func (m MockCluster) WithVerifiedSource(ctx context.Context) context.Context {
	if m.WithVerifiedSourceFunc == nil {
		return ctx
	}
	return m.WithVerifiedSourceFunc(ctx)
}

type GetPeerRequest struct {
	Context     context.Context
	Role        ttnpb.PeerInfo_Role
	Identifiers ttnpb.Identifiers
	Response    chan<- cluster.Peer
}

func MakeGetPeerChFunc(reqCh chan<- GetPeerRequest) func(context.Context, ttnpb.PeerInfo_Role, ttnpb.Identifiers) cluster.Peer {
	return func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer {
		respCh := make(chan cluster.Peer)
		reqCh <- GetPeerRequest{
			Context:     ctx,
			Role:        role,
			Identifiers: ids,
			Response:    respCh,
		}
		return <-respCh
	}
}

func AssertGetPeerRequest(t *testing.T, reqCh <-chan GetPeerRequest, timeout time.Duration, assert func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) bool, peer cluster.Peer) bool {
	t.Helper()
	select {
	case req := <-reqCh:
		if !assert(req.Context, req.Role, req.Identifiers) {
			return false
		}
		select {
		case req.Response <- peer:
			return true

		case <-time.After(timeout):
			t.Error("Timed out while waiting for cluster.GetPeer response to be processed")
			return false
		}

	case <-time.After(timeout):
		t.Error("Timed out while waiting for cluster.GetPeer request to arrive")
		return false
	}
}
