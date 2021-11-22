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
	return m.GetIds().EntityType()
}

func (m *GetApplicationRequest) EntityType() string {
	return m.GetApplicationIds().EntityType()
}

func (m *CreateApplicationRequest) EntityType() string {
	return m.GetApplication().EntityType()
}

func (m *UpdateApplicationRequest) EntityType() string {
	return m.GetApplication().EntityType()
}

func (m *ListApplicationAPIKeysRequest) EntityType() string {
	return m.GetApplicationIds().EntityType()
}

func (m *GetApplicationAPIKeyRequest) EntityType() string {
	return m.GetApplicationIds().EntityType()
}

func (m *CreateApplicationAPIKeyRequest) EntityType() string {
	return m.GetApplicationIds().EntityType()
}

func (m *UpdateApplicationAPIKeyRequest) EntityType() string {
	return m.GetApplicationIds().EntityType()
}

func (m *ListApplicationCollaboratorsRequest) EntityType() string {
	return m.GetApplicationIds().EntityType()
}

func (m *GetApplicationCollaboratorRequest) EntityType() string {
	return m.GetApplicationIds().EntityType()
}

func (m *SetApplicationCollaboratorRequest) EntityType() string {
	return m.GetApplicationIds().EntityType()
}

// All IDString methods implement the IDStringer interface.

func (m *Application) IDString() string {
	return m.GetIds().IDString()
}

func (m *GetApplicationRequest) IDString() string {
	return m.GetApplicationIds().IDString()
}

func (m *CreateApplicationRequest) IDString() string {
	return m.GetApplication().IDString()
}

func (m *UpdateApplicationRequest) IDString() string {
	return m.GetApplication().IDString()
}

func (m *ListApplicationAPIKeysRequest) IDString() string {
	return m.GetApplicationIds().IDString()
}

func (m *GetApplicationAPIKeyRequest) IDString() string {
	return m.GetApplicationIds().IDString()
}

func (m *CreateApplicationAPIKeyRequest) IDString() string {
	return m.GetApplicationIds().IDString()
}

func (m *UpdateApplicationAPIKeyRequest) IDString() string {
	return m.GetApplicationIds().IDString()
}

func (m *ListApplicationCollaboratorsRequest) IDString() string {
	return m.GetApplicationIds().IDString()
}

func (m *GetApplicationCollaboratorRequest) IDString() string {
	return m.GetApplicationIds().IDString()
}

func (m *SetApplicationCollaboratorRequest) IDString() string {
	return m.GetApplicationIds().IDString()
}

// All ExtractRequestFields methods are used by github.com/grpc-ecosystem/go-grpc-middleware/tags.

func (m *Application) ExtractRequestFields(dst map[string]interface{}) {
	m.GetIds().ExtractRequestFields(dst)
}

func (m *GetApplicationRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetApplicationIds().ExtractRequestFields(dst)
}

func (m *UpdateApplicationRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetApplication().ExtractRequestFields(dst)
}

func (m *ListApplicationAPIKeysRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetApplicationIds().ExtractRequestFields(dst)
}

func (m *GetApplicationAPIKeyRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetApplicationIds().ExtractRequestFields(dst)
}

func (m *CreateApplicationAPIKeyRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetApplicationIds().ExtractRequestFields(dst)
}

func (m *UpdateApplicationAPIKeyRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetApplicationIds().ExtractRequestFields(dst)
}

func (m *ListApplicationCollaboratorsRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetApplicationIds().ExtractRequestFields(dst)
}

func (m *GetApplicationCollaboratorRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetApplicationIds().ExtractRequestFields(dst)
}

func (m *ApplicationUp) ExtractRequestFields(dst map[string]interface{}) {
	ids := m.EndDeviceIds
	if ids == nil {
		return
	}
	ids.ExtractRequestFields(dst)
}

func (m *DownlinkQueueRequest) ExtractRequestFields(dst map[string]interface{}) {
	ids := m.EndDeviceIds
	if ids == nil {
		return
	}
	ids.ExtractRequestFields(dst)
}

// Wrap methods of m.ApplicationIdentifiers.

// GetEntityIdentifiers returns entity identifiers.
func (m *Application) GetEntityIdentifiers() *EntityIdentifiers {
	if m == nil {
		return nil
	}
	return m.GetIds().GetEntityIdentifiers()
}

// EntityType implements IDStringer.
func (m *ApplicationUp) EntityType() string {
	ids := m.EndDeviceIds
	if ids == nil {
		return ""
	}
	return ids.EntityType()
}

// IDString implements IDStringer.
func (m *ApplicationUp) IDString() string {
	ids := m.EndDeviceIds
	if ids == nil {
		return ""
	}
	return ids.IDString()
}

// EntityType implements IDStringer.
func (m *DownlinkQueueRequest) EntityType() string {
	ids := m.EndDeviceIds
	if ids == nil {
		return ""
	}
	return ids.EntityType()
}

// IDString implements IDStringer.
func (m *DownlinkQueueRequest) IDString() string {
	ids := m.EndDeviceIds
	if ids == nil {
		return ""
	}
	return ids.IDString()
}
