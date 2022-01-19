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

package ttnpb

// GetEntityIdentifiers returns the EntityIdentifiers for the used access method.
func (m *AuthInfoResponse) GetEntityIdentifiers() *EntityIdentifiers {
	if m == nil {
		return nil
	}
	switch accessMethod := m.GetAccessMethod().(type) {
	case *AuthInfoResponse_ApiKey:
		return accessMethod.ApiKey.EntityIds
	case *AuthInfoResponse_OauthAccessToken:
		return accessMethod.OauthAccessToken.UserIds.GetEntityIdentifiers()
	case *AuthInfoResponse_UserSession:
		return accessMethod.UserSession.GetUserIds().GetEntityIdentifiers()
	}
	return nil
}

// GetRights returns the entity Rights for the used access method.
func (m *AuthInfoResponse) GetRights() []Right {
	if m == nil {
		return nil
	}
	switch accessMethod := m.GetAccessMethod().(type) {
	case *AuthInfoResponse_ApiKey:
		return accessMethod.ApiKey.GetApiKey().GetRights()
	case *AuthInfoResponse_OauthAccessToken:
		return accessMethod.OauthAccessToken.GetRights()
	case *AuthInfoResponse_UserSession:
		return RightsFrom(Right_RIGHT_ALL).Implied().GetRights()
	}
	return nil
}

// GetOrganizationOrUserIdentifiers returns the OrganizationOrUserIdentifiers for the used access method.
func (m *AuthInfoResponse) GetOrganizationOrUserIdentifiers() *OrganizationOrUserIdentifiers {
	ids := m.GetEntityIdentifiers()
	if ids == nil {
		return nil
	}
	if ids := ids.GetOrganizationIds(); ids != nil {
		return ids.OrganizationOrUserIdentifiers()
	}
	if ids := ids.GetUserIds(); ids != nil {
		return ids.OrganizationOrUserIdentifiers()
	}
	return nil
}
