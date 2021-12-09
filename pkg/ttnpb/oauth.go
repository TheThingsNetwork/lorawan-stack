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

func (m *OAuthClientAuthorizationIdentifiers) EntityType() string {
	return m.GetUserIds().EntityType()
}

func (m *OAuthClientAuthorization) EntityType() string {
	return m.GetUserIds().EntityType()
}

func (m *ListOAuthClientAuthorizationsRequest) EntityType() string {
	return m.GetUserIds().EntityType()
}

func (m *OAuthAuthorizationCode) EntityType() string {
	return m.GetUserIds().EntityType()
}

func (m *OAuthAccessTokenIdentifiers) EntityType() string {
	return m.GetUserIds().EntityType()
}

func (m *OAuthAccessToken) EntityType() string {
	return m.GetUserIds().EntityType()
}

func (m *ListOAuthAccessTokensRequest) EntityType() string {
	return m.GetUserIds().EntityType()
}

// All IDString methods implement the IDStringer interface.

func (m *OAuthClientAuthorizationIdentifiers) IDString() string {
	return m.GetUserIds().IDString()
}

func (m *OAuthClientAuthorization) IDString() string {
	return m.GetUserIds().IDString()
}

func (m *ListOAuthClientAuthorizationsRequest) IDString() string {
	return m.GetUserIds().IDString()
}

func (m *OAuthAuthorizationCode) IDString() string {
	return m.GetUserIds().IDString()
}

func (m *OAuthAccessTokenIdentifiers) IDString() string {
	return m.GetUserIds().IDString()
}

func (m *OAuthAccessToken) IDString() string {
	return m.GetUserIds().IDString()
}

func (m *ListOAuthAccessTokensRequest) IDString() string {
	return m.GetUserIds().IDString()
}

// All ExtractRequestFields methods are used by github.com/grpc-ecosystem/go-grpc-middleware/tags.

func (m *OAuthClientAuthorizationIdentifiers) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

func (m *OAuthClientAuthorization) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

func (m *ListOAuthClientAuthorizationsRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

func (m *OAuthAuthorizationCode) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

func (m *OAuthAccessTokenIdentifiers) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

func (m *OAuthAccessToken) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}

func (m *ListOAuthAccessTokensRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetUserIds().ExtractRequestFields(dst)
}
