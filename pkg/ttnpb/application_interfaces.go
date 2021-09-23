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

func (m *Application) EntityType() string {
	return m.Ids.EntityType()
}

func (m *GetApplicationRequest) EntityType() string {
	return m.ApplicationIds.EntityType()
}

func (m *ListApplicationAPIKeysRequest) EntityType() string {
	return m.ApplicationIds.EntityType()
}

func (m *GetApplicationAPIKeyRequest) EntityType() string {
	return m.ApplicationIds.EntityType()
}

func (m *CreateApplicationAPIKeyRequest) EntityType() string {
	return m.ApplicationIds.EntityType()
}

func (m *UpdateApplicationAPIKeyRequest) EntityType() string {
	return m.ApplicationIds.EntityType()
}

func (m *ListApplicationCollaboratorsRequest) EntityType() string {
	return m.ApplicationIds.EntityType()
}

func (m *GetApplicationCollaboratorRequest) EntityType() string {
	return m.ApplicationIds.EntityType()
}

func (m *SetApplicationCollaboratorRequest) EntityType() string {
	return m.ApplicationIds.EntityType()
}

// All IDString methods implement the IDStringer interface.

func (m *Application) IDString() string {
	return m.Ids.IDString()
}

func (m *GetApplicationRequest) IDString() string {
	return m.ApplicationIds.IDString()
}

func (m *ListApplicationAPIKeysRequest) IDString() string {
	return m.ApplicationIds.IDString()
}

func (m *GetApplicationAPIKeyRequest) IDString() string {
	return m.ApplicationIds.IDString()
}

func (m *CreateApplicationAPIKeyRequest) IDString() string {
	return m.ApplicationIds.IDString()
}

func (m *UpdateApplicationAPIKeyRequest) IDString() string {
	return m.ApplicationIds.IDString()
}

func (m *ListApplicationCollaboratorsRequest) IDString() string {
	return m.ApplicationIds.IDString()
}

func (m *GetApplicationCollaboratorRequest) IDString() string {
	return m.ApplicationIds.IDString()
}

func (m *SetApplicationCollaboratorRequest) IDString() string {
	return m.ApplicationIds.IDString()
}

// All ExtractRequestFields methods are used by github.com/grpc-ecosystem/go-grpc-middleware/tags.

func (m *Application) ExtractRequestFields(dst map[string]interface{}) {
	m.Ids.ExtractRequestFields(dst)
}

func (m *GetApplicationRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.ApplicationIds.ExtractRequestFields(dst)
}

func (m *ListApplicationAPIKeysRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.ApplicationIds.ExtractRequestFields(dst)
}

func (m *GetApplicationAPIKeyRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.ApplicationIds.ExtractRequestFields(dst)
}

func (m *CreateApplicationAPIKeyRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.ApplicationIds.ExtractRequestFields(dst)
}

func (m *UpdateApplicationAPIKeyRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.ApplicationIds.ExtractRequestFields(dst)
}

func (m *ListApplicationCollaboratorsRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.ApplicationIds.ExtractRequestFields(dst)
}

func (m *GetApplicationCollaboratorRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.ApplicationIds.ExtractRequestFields(dst)
}

// Wrap methods of m.ApplicationIdentifiers.

func (m *Application) GetEntityIdentifiers() *EntityIdentifiers {
	if m == nil {
		return nil
	}
	return m.Ids.GetEntityIdentifiers()
}
