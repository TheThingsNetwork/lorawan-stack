// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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

package ttnpb_test

import (
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func TestCreateOrganizationRequest_RateLimitKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		req      *ttnpb.CreateOrganizationRequest
		expected string
	}{
		{
			name: "valid request with user collaborator",
			req: &ttnpb.CreateOrganizationRequest{
				Organization: &ttnpb.Organization{
					Ids: &ttnpb.OrganizationIdentifiers{
						OrganizationId: "test-org-1",
					},
				},
				Collaborator: &ttnpb.OrganizationOrUserIdentifiers{
					Ids: &ttnpb.OrganizationOrUserIdentifiers_UserIds{
						UserIds: &ttnpb.UserIdentifiers{
							UserId: "test-user",
						},
					},
				},
			},
			expected: "",
		},
		{
			name: "different organization same user",
			req: &ttnpb.CreateOrganizationRequest{
				Organization: &ttnpb.Organization{
					Ids: &ttnpb.OrganizationIdentifiers{
						OrganizationId: "test-org-2",
					},
				},
				Collaborator: &ttnpb.OrganizationOrUserIdentifiers{
					Ids: &ttnpb.OrganizationOrUserIdentifiers_UserIds{
						UserIds: &ttnpb.UserIdentifiers{
							UserId: "test-user",
						},
					},
				},
			},
			expected: "",
		},
		{
			name: "nil collaborator",
			req: &ttnpb.CreateOrganizationRequest{
				Organization: &ttnpb.Organization{
					Ids: &ttnpb.OrganizationIdentifiers{
						OrganizationId: "test-org",
					},
				},
			},
			expected: "",
		},
		{
			name:     "nil request",
			req:      &ttnpb.CreateOrganizationRequest{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.req.RateLimitKey()
			if got != tt.expected {
				t.Errorf("RateLimitKey() = %v, want %v", got, tt.expected)
			}
		})
	}
}
