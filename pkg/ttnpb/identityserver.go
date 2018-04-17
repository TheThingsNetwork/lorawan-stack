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

import "time"

// IsIDAllowed checks whether an ID is allowed to be used or not given the list
// of blacklisted IDs of the receiver.
func (s *IdentityServerSettings) IsIDAllowed(id string) bool {
	for _, blacklistedID := range s.BlacklistedIDs {
		if blacklistedID == id {
			return false
		}
	}
	return true
}

// IsExpired checks whether or not the invitation is expired.
func (i *ListInvitationsResponse_Invitation) IsExpired() bool {
	return i.ExpiresAt.Before(time.Now())
}
