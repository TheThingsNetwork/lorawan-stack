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

package ttnpb

// All EntityType methods implement the IDStringer interface.

func (m *Gateway) EntityType() string {
	return m.GetIds().EntityType()
}

func (m *GetGatewayRequest) EntityType() string {
	return m.GetGatewayIds().EntityType()
}

func (m *CreateGatewayRequest) EntityType() string {
	return m.GetGateway().EntityType()
}

func (m *UpdateGatewayRequest) EntityType() string {
	return m.GetGateway().EntityType()
}

func (m *ListGatewayAPIKeysRequest) EntityType() string {
	return m.GetGatewayIds().EntityType()
}

func (m *GetGatewayAPIKeyRequest) EntityType() string {
	return m.GetGatewayIds().EntityType()
}

func (m *CreateGatewayAPIKeyRequest) EntityType() string {
	return m.GetGatewayIds().EntityType()
}

func (m *UpdateGatewayAPIKeyRequest) EntityType() string {
	return m.GetGatewayIds().EntityType()
}

func (m *ListGatewayCollaboratorsRequest) EntityType() string {
	return m.GetGatewayIds().EntityType()
}

func (m *GetGatewayCollaboratorRequest) EntityType() string {
	return m.GetGatewayIds().EntityType()
}

func (m *SetGatewayCollaboratorRequest) EntityType() string {
	return m.GetGatewayIds().EntityType()
}

// All IDString methods implement the IDStringer interface.

func (m *Gateway) IDString() string {
	return m.GetIds().IDString()
}

func (m *GetGatewayRequest) IDString() string {
	return m.GetGatewayIds().IDString()
}

func (m *CreateGatewayRequest) IDString() string {
	return m.GetGateway().IDString()
}

func (m *UpdateGatewayRequest) IDString() string {
	return m.GetGateway().IDString()
}

func (m *ListGatewayAPIKeysRequest) IDString() string {
	return m.GetGatewayIds().IDString()
}

func (m *GetGatewayAPIKeyRequest) IDString() string {
	return m.GetGatewayIds().IDString()
}

func (m *CreateGatewayAPIKeyRequest) IDString() string {
	return m.GetGatewayIds().IDString()
}

func (m *UpdateGatewayAPIKeyRequest) IDString() string {
	return m.GetGatewayIds().IDString()
}

func (m *ListGatewayCollaboratorsRequest) IDString() string {
	return m.GetGatewayIds().IDString()
}

func (m *GetGatewayCollaboratorRequest) IDString() string {
	return m.GetGatewayIds().IDString()
}

func (m *SetGatewayCollaboratorRequest) IDString() string {
	return m.GetGatewayIds().IDString()
}

// All ExtractRequestFields methods are used by github.com/grpc-ecosystem/go-grpc-middleware/tags.

func (m *Gateway) ExtractRequestFields(dst map[string]interface{}) {
	m.GetIds().ExtractRequestFields(dst)
}

func (m *GetGatewayRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetGatewayIds().ExtractRequestFields(dst)
}

func (m *UpdateGatewayRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetGateway().ExtractRequestFields(dst)
}

func (m *ListGatewayAPIKeysRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetGatewayIds().ExtractRequestFields(dst)
}

func (m *GetGatewayAPIKeyRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetGatewayIds().ExtractRequestFields(dst)
}

func (m *CreateGatewayAPIKeyRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetGatewayIds().ExtractRequestFields(dst)
}

func (m *UpdateGatewayAPIKeyRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetGatewayIds().ExtractRequestFields(dst)
}

func (m *ListGatewayCollaboratorsRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetGatewayIds().ExtractRequestFields(dst)
}

func (m *GetGatewayCollaboratorRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetGatewayIds().ExtractRequestFields(dst)
}

// Wrap methods of m.GatewayIdentifiers.

func (m *Gateway) GetEntityIdentifiers() *EntityIdentifiers {
	if m == nil {
		return nil
	}
	return m.GetIds().GetEntityIdentifiers()
}
