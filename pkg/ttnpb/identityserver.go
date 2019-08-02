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
	case *AuthInfoResponse_APIKey:
		return &accessMethod.APIKey.EntityIDs
	case *AuthInfoResponse_OAuthAccessToken:
		return accessMethod.OAuthAccessToken.UserIDs.EntityIdentifiers()
	}
	return nil
}

// GetOrganizationOrUserIdentifiers returns the OrganizationOrUserIdentifiers for the used access method.
func (m *AuthInfoResponse) GetOrganizationOrUserIdentifiers() *OrganizationOrUserIdentifiers {
	ids := m.GetEntityIdentifiers()
	if ids == nil {
		return nil
	}
	if ids := ids.GetOrganizationIDs(); ids != nil {
		return ids.OrganizationOrUserIdentifiers()
	}
	if ids := ids.GetUserIDs(); ids != nil {
		return ids.OrganizationOrUserIdentifiers()
	}
	return nil
}
