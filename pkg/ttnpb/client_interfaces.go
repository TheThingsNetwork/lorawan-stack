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

func (m *Client) EntityType() string {
	return m.GetIds().EntityType()
}

func (m *GetClientRequest) EntityType() string {
	return m.GetClientIds().EntityType()
}

func (m *CreateClientRequest) EntityType() string {
	return m.GetClient().EntityType()
}

func (m *UpdateClientRequest) EntityType() string {
	return m.GetClient().EntityType()
}

func (m *ListClientCollaboratorsRequest) EntityType() string {
	return m.GetClientIds().EntityType()
}

func (m *GetClientCollaboratorRequest) EntityType() string {
	return m.GetClientIds().EntityType()
}

func (m *SetClientCollaboratorRequest) EntityType() string {
	return m.GetClientIds().EntityType()
}

// All IDString methods implement the IDStringer interface.

func (m *Client) IDString() string {
	return m.GetIds().IDString()
}

func (m *GetClientRequest) IDString() string {
	return m.GetClientIds().IDString()
}

func (m *CreateClientRequest) IDString() string {
	return m.GetClient().IDString()
}

func (m *UpdateClientRequest) IDString() string {
	return m.GetClient().IDString()
}

func (m *ListClientCollaboratorsRequest) IDString() string {
	return m.GetClientIds().IDString()
}

func (m *GetClientCollaboratorRequest) IDString() string {
	return m.GetClientIds().IDString()
}

func (m *SetClientCollaboratorRequest) IDString() string {
	return m.GetClientIds().IDString()
}

// All ExtractRequestFields methods are used by github.com/grpc-ecosystem/go-grpc-middleware/tags.

func (m *Client) ExtractRequestFields(dst map[string]interface{}) {
	m.GetIds().ExtractRequestFields(dst)
}

func (m *GetClientRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetClientIds().ExtractRequestFields(dst)
}

func (m *UpdateClientRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetClient().ExtractRequestFields(dst)
}

func (m *ListClientCollaboratorsRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetClientIds().ExtractRequestFields(dst)
}

func (m *GetClientCollaboratorRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetClientIds().ExtractRequestFields(dst)
}

// Wrap methods of m.ClientIdentifiers.

func (m *Client) GetEntityIdentifiers() *EntityIdentifiers {
	if m == nil {
		return nil
	}
	return m.GetIds().GetEntityIdentifiers()
}
