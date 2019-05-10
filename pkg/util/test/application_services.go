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
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// MockApplicationAccessServer is a mock ttnpb.ApplicationAccessServer used for testing.
type MockApplicationAccessServer struct {
	CreateAPIKeyFunc      func(context.Context, *ttnpb.CreateApplicationAPIKeyRequest) (*ttnpb.APIKey, error)
	GetAPIKeyFunc         func(context.Context, *ttnpb.GetApplicationAPIKeyRequest) (*ttnpb.APIKey, error)
	ListAPIKeysFunc       func(context.Context, *ttnpb.ListApplicationAPIKeysRequest) (*ttnpb.APIKeys, error)
	ListCollaboratorsFunc func(context.Context, *ttnpb.ListApplicationCollaboratorsRequest) (*ttnpb.Collaborators, error)
	ListRightsFunc        func(context.Context, *ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error)
	SetCollaboratorFunc   func(context.Context, *ttnpb.SetApplicationCollaboratorRequest) (*pbtypes.Empty, error)
	UpdateAPIKeyFunc      func(context.Context, *ttnpb.UpdateApplicationAPIKeyRequest) (*ttnpb.APIKey, error)
}

// ListRights calls ListRightsFunc if set and returns zero value otherwise.
func (m MockApplicationAccessServer) ListRights(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	if m.ListRightsFunc == nil {
		return &ttnpb.Rights{}, nil
	}
	return m.ListRightsFunc(ctx, req)
}

// CreateAPIKey calls CreateAPIKeyFunc if set and returns zero value otherwise.
func (m MockApplicationAccessServer) CreateAPIKey(ctx context.Context, req *ttnpb.CreateApplicationAPIKeyRequest) (*ttnpb.APIKey, error) {
	if m.CreateAPIKeyFunc == nil {
		return &ttnpb.APIKey{}, nil
	}
	return m.CreateAPIKeyFunc(ctx, req)
}

// ListAPIKeys calls ListAPIKeysFunc if set and returns zero value otherwise.
func (m MockApplicationAccessServer) ListAPIKeys(ctx context.Context, req *ttnpb.ListApplicationAPIKeysRequest) (*ttnpb.APIKeys, error) {
	if m.ListAPIKeysFunc == nil {
		return &ttnpb.APIKeys{}, nil
	}
	return m.ListAPIKeysFunc(ctx, req)
}

// GetAPIKey calls GetAPIKeyFunc if set and returns zero value otherwise.
func (m MockApplicationAccessServer) GetAPIKey(ctx context.Context, req *ttnpb.GetApplicationAPIKeyRequest) (*ttnpb.APIKey, error) {
	if m.GetAPIKeyFunc == nil {
		return &ttnpb.APIKey{}, nil
	}
	return m.GetAPIKeyFunc(ctx, req)
}

// UpdateAPIKey calls UpdateAPIKeyFunc if set and returns zero value otherwise.
func (m MockApplicationAccessServer) UpdateAPIKey(ctx context.Context, req *ttnpb.UpdateApplicationAPIKeyRequest) (*ttnpb.APIKey, error) {
	if m.UpdateAPIKeyFunc == nil {
		return &ttnpb.APIKey{}, nil
	}
	return m.UpdateAPIKeyFunc(ctx, req)
}

// SetCollaborator calls SetCollaboratorFunc if set and returns zero value otherwise.
func (m MockApplicationAccessServer) SetCollaborator(ctx context.Context, req *ttnpb.SetApplicationCollaboratorRequest) (*pbtypes.Empty, error) {
	if m.SetCollaboratorFunc == nil {
		return &pbtypes.Empty{}, nil
	}
	return m.SetCollaboratorFunc(ctx, req)
}

// ListCollaborators calls ListCollaboratorsFunc if set and returns zero value otherwise.
func (m MockApplicationAccessServer) ListCollaborators(ctx context.Context, req *ttnpb.ListApplicationCollaboratorsRequest) (*ttnpb.Collaborators, error) {
	if m.ListCollaboratorsFunc == nil {
		return &ttnpb.Collaborators{}, nil
	}
	return m.ListCollaboratorsFunc(ctx, req)
}

type ApplicationAccessListRightsResponse struct {
	Response *ttnpb.Rights
	Error    error
}
type ApplicationAccessListRightsRequest struct {
	Context  context.Context
	Message  *ttnpb.ApplicationIdentifiers
	Response chan<- ApplicationAccessListRightsResponse
}

func MakeApplicationAccessListRightsChFunc(reqCh chan<- ApplicationAccessListRightsRequest) func(context.Context, *ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	return func(ctx context.Context, msg *ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
		respCh := make(chan ApplicationAccessListRightsResponse)
		reqCh <- ApplicationAccessListRightsRequest{
			Context:  ctx,
			Message:  msg,
			Response: respCh,
		}
		resp := <-respCh
		return resp.Response, resp.Error
	}
}

func AssertListRightsRequest(t *testing.T, reqCh <-chan ApplicationAccessListRightsRequest, timeout time.Duration, assert func(ctx context.Context, ids ttnpb.Identifiers) bool, rights ...ttnpb.Right) bool {
	t.Helper()
	select {
	case req := <-reqCh:
		if !assert(req.Context, req.Message) {
			return false
		}
		select {
		case req.Response <- ApplicationAccessListRightsResponse{
			Response: &ttnpb.Rights{
				Rights: rights,
			},
		}:
			return true

		case <-time.After(timeout):
			t.Error("Timed out while waiting for ApplicationAccess.ListRights response to be processed")
			return false
		}

	case <-time.After(timeout):
		t.Error("Timed out while waiting for ApplicationAccess.ListRights request to arrive")
		return false
	}
}
