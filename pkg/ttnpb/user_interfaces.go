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

func (m *User) EntityType() string {
	return m.GetIds().EntityType()
}

func (m *GetUserRequest) EntityType() string {
	return m.GetUserIds().EntityType()
}

func (m *CreateUserRequest) EntityType() string {
	return m.GetUser().GetIds().EntityType()
}

func (m *UpdateUserRequest) EntityType() string {
	return m.GetUser().GetIds().EntityType()
}

func (m *CreateTemporaryPasswordRequest) EntityType() string {
	return m.UserIds.EntityType()
}

func (m *UpdateUserPasswordRequest) EntityType() string {
	return m.GetUserIds().EntityType()
}

func (m *ListUserAPIKeysRequest) EntityType() string {
	return m.GetUserIds().EntityType()
}

func (m *GetUserAPIKeyRequest) EntityType() string {
	return m.GetUserIds().EntityType()
}

func (m *CreateUserAPIKeyRequest) EntityType() string {
	return m.GetUserIds().EntityType()
}

func (m *UpdateUserAPIKeyRequest) EntityType() string {
	return m.GetUserIds().EntityType()
}

func (m *UserSessionIdentifiers) EntityType() string {
	return m.GetUserIds().EntityType()
}

func (m *UserSession) EntityType() string {
	return m.GetUserIds().EntityType()
}

func (m *ListUserSessionsRequest) EntityType() string {
	return m.GetUserIds().EntityType()
}

func (m *LoginToken) EntityType() string {
	return m.GetUserIds().EntityType()
}

func (m *CreateLoginTokenRequest) EntityType() string {
	return m.GetUserIds().EntityType()
}

// All IDString methods implement the IDStringer interface.

func (m *User) IDString() string {
	return m.GetIds().IDString()
}

func (m *GetUserRequest) IDString() string {
	return m.GetUserIds().IDString()
}

func (m *CreateUserRequest) IDString() string {
	return m.GetUser().GetIds().IDString()
}

func (m *UpdateUserRequest) IDString() string {
	return m.GetUser().GetIds().IDString()
}

func (m *CreateTemporaryPasswordRequest) IDString() string {
	return m.UserIds.IDString()
}

func (m *UpdateUserPasswordRequest) IDString() string {
	return m.GetUserIds().IDString()
}

func (m *ListUserAPIKeysRequest) IDString() string {
	return m.GetUserIds().IDString()
}

func (m *GetUserAPIKeyRequest) IDString() string {
	return m.GetUserIds().IDString()
}

func (m *CreateUserAPIKeyRequest) IDString() string {
	return m.GetUserIds().IDString()
}

func (m *UpdateUserAPIKeyRequest) IDString() string {
	return m.GetUserIds().IDString()
}

func (m *UserSessionIdentifiers) IDString() string {
	return m.GetUserIds().IDString()
}

func (m *UserSession) IDString() string {
	return m.GetUserIds().IDString()
}

func (m *ListUserSessionsRequest) IDString() string {
	return m.GetUserIds().IDString()
}

func (m *LoginToken) IDString() string {
	return m.GetUserIds().IDString()
}

func (m *CreateLoginTokenRequest) IDString() string {
	return m.GetUserIds().IDString()
}

// All ExtractRequestFields methods are used by github.com/grpc-ecosystem/go-grpc-middleware/tags.

func (m *GetUserRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

func (m *CreateUserRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUser().GetIds().ExtractRequestFields(dst)
}

func (m *UpdateUserRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUser().GetIds().ExtractRequestFields(dst)
}

func (m *CreateTemporaryPasswordRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.UserIds.ExtractRequestFields(dst)
}

func (m *UpdateUserPasswordRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

func (m *ListUserAPIKeysRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

func (m *GetUserAPIKeyRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

func (m *CreateUserAPIKeyRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

func (m *UpdateUserAPIKeyRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

func (m *UserSessionIdentifiers) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

func (m *UserSession) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

func (m *ListUserSessionsRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

func (m *LoginToken) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

func (m *CreateLoginTokenRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

// Wrap methods of m.UserIdentifiers.

func (m *User) GetEntityIdentifiers() *EntityIdentifiers {
	return m.GetIds().GetEntityIdentifiers()
}

func (m *User) GetOrganizationOrUserIdentifiers() *OrganizationOrUserIdentifiers {
	return m.GetIds().GetOrganizationOrUserIdentifiers()
}
