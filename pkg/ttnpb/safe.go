// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

// PublicSafe returns a copy of the client with only the fields that are safe to
// return to any audience.
func (c Client) PublicSafe() Client {
	var safe Client
	safe.ClientIdentifiers = c.ClientIdentifiers
	safe.CreatedAt = c.CreatedAt
	safe.UpdatedAt = c.UpdatedAt
	safe.Name = c.Name
	safe.Description = c.Description
	if len(c.ContactInfo) > 0 {
		safe.ContactInfo = make([]*ContactInfo, 0, len(c.ContactInfo))
		for _, info := range c.ContactInfo {
			if !info.Public {
				continue
			}
			safe.ContactInfo = append(safe.ContactInfo, info)
		}
	}
	safe.RedirectURIs = c.RedirectURIs
	safe.State = c.State
	safe.SkipAuthorization = c.SkipAuthorization
	safe.Endorsed = c.Endorsed
	safe.Grants = c.Grants
	safe.Rights = c.Rights
	return safe
}

// PublicSafe returns a copy of the user with only the fields that are safe to
// return to any audience.
func (u User) PublicSafe() User {
	var safe User
	safe.UserIdentifiers = u.UserIdentifiers
	safe.CreatedAt = u.CreatedAt
	safe.UpdatedAt = u.UpdatedAt
	safe.Name = u.Name
	safe.Description = u.Description
	if len(u.ContactInfo) > 0 {
		safe.ContactInfo = make([]*ContactInfo, 0, len(u.ContactInfo))
		for _, info := range u.ContactInfo {
			if !info.Public {
				continue
			}
			safe.ContactInfo = append(safe.ContactInfo, info)
		}
	}
	safe.State = u.State
	safe.Admin = u.Admin
	safe.ProfilePicture = u.ProfilePicture
	return safe
}
