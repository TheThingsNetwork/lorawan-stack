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
	return m.Ids.EntityType()
}

func (m *GetClientRequest) EntityType() string {
	return m.ClientIds.EntityType()
}

func (m *ListClientCollaboratorsRequest) EntityType() string {
	return m.ClientIds.EntityType()
}

func (m *GetClientCollaboratorRequest) EntityType() string {
	return m.ClientIds.EntityType()
}

func (m *SetClientCollaboratorRequest) EntityType() string {
	return m.ClientIds.EntityType()
}

// All IDString methods implement the IDStringer interface.

func (m *Client) IDString() string {
	return m.Ids.IDString()
}

func (m *GetClientRequest) IDString() string {
	return m.ClientIds.IDString()
}

func (m *ListClientCollaboratorsRequest) IDString() string {
	return m.ClientIds.IDString()
}

func (m *GetClientCollaboratorRequest) IDString() string {
	return m.ClientIds.IDString()
}

func (m *SetClientCollaboratorRequest) IDString() string {
	return m.ClientIds.IDString()
}

// All ExtractRequestFields methods are used by github.com/grpc-ecosystem/go-grpc-middleware/tags.

func (m *Client) ExtractRequestFields(dst map[string]interface{}) {
	m.Ids.ExtractRequestFields(dst)
}

func (m *GetClientRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.ClientIds.ExtractRequestFields(dst)
}

func (m *ListClientCollaboratorsRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.ClientIds.ExtractRequestFields(dst)
}

func (m *GetClientCollaboratorRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.ClientIds.ExtractRequestFields(dst)
}

// Wrap methods of m.ClientIdentifiers.

func (m *Client) GetEntityIdentifiers() *EntityIdentifiers {
	if m == nil {
		return nil
	}
	return m.Ids.GetEntityIdentifiers()
}
