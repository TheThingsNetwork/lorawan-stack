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

func onlyPublicContactInfo(info []*ContactInfo) []*ContactInfo {
	if info == nil {
		return nil
	}
	out := make([]*ContactInfo, 0, len(info))
	for _, info := range info {
		if !info.Public {
			continue
		}
		out = append(out, info)
	}
	return out
}

// PublicEntityFields are the fields that are public for each entity.
var PublicEntityFields = []string{
	"ids",
	"created_at",
	"updated_at",
	"deleted_at",
	"contact_info", // Note that this is filtered.
}

// PublicApplicationFields are the Application's fields that are public.
var PublicApplicationFields = append(PublicEntityFields)

// PublicSafe returns a copy of the application with only the fields that are
// safe to return to any audience.
func (a *Application) PublicSafe() *Application {
	if a == nil {
		return nil
	}
	var safe Application
	safe.SetFields(a, PublicApplicationFields...)
	safe.ContactInfo = onlyPublicContactInfo(safe.ContactInfo)
	return &safe
}

// PublicClientFields are the Client's fields that are public.
var PublicClientFields = append(PublicEntityFields,
	"name",
	"description",
	"redirect_uris",
	"logout_redirect_uris",
	"state",
	"skip_authorization",
	"endorsed",
	"grants",
	"rights",
)

// PublicSafe returns a copy of the client with only the fields that are safe to
// return to any audience.
func (c *Client) PublicSafe() *Client {
	if c == nil {
		return nil
	}
	var safe Client
	safe.SetFields(c, PublicClientFields...)
	safe.ContactInfo = onlyPublicContactInfo(safe.ContactInfo)
	return &safe
}

// PublicGatewayFields are the Gateway's fields that are public.
var PublicGatewayFields = append(PublicEntityFields,
	"name",
	"description",
	"frequency_plan_id",
	"frequency_plan_ids",
	"status_public",
	"gateway_server_address", // only public if status_public=true
	"location_public",
	"antennas", // only public if location_public=true
	"lrfhss.supported",
)

// PublicSafe returns a copy of the gateway with only the fields that are
// safe to return to any audience.
func (g *Gateway) PublicSafe() *Gateway {
	if g == nil {
		return nil
	}
	var safe Gateway
	safe.SetFields(g, PublicGatewayFields...)
	safe.ContactInfo = onlyPublicContactInfo(safe.ContactInfo)
	if !safe.StatusPublic {
		safe.GatewayServerAddress = ""
	}
	if !safe.LocationPublic {
		for _, ant := range safe.Antennas {
			ant.Location = nil
		}
	}
	return &safe
}

// PublicOrganizationFields are the Organization's fields that are public.
var PublicOrganizationFields = append(PublicEntityFields,
	"name",
)

// PublicSafe returns a copy of the organization with only the fields that are
// safe to return to any audience.
func (o *Organization) PublicSafe() *Organization {
	if o == nil {
		return nil
	}
	var safe Organization
	safe.SetFields(o, PublicOrganizationFields...)
	safe.ContactInfo = onlyPublicContactInfo(safe.ContactInfo)
	return &safe
}

// PublicUserFields are the User's fields that are public.
var PublicUserFields = append(PublicEntityFields,
	"name",
	"description",
	"state",
	"admin",
	"profile_picture",
)

// PublicSafe returns a copy of the user with only the fields that are safe to
// return to any audience.
func (u *User) PublicSafe() *User {
	if u == nil {
		return nil
	}
	var safe User
	safe.SetFields(u, PublicUserFields...)
	safe.ContactInfo = onlyPublicContactInfo(safe.ContactInfo)
	return &safe
}

// PublicSafe returns only the identifiers of the collaborators.
func (c *Collaborators) PublicSafe() *Collaborators {
	if c == nil {
		return nil
	}
	safe := Collaborators{
		Collaborators: make([]*Collaborator, len(c.Collaborators)),
	}
	for i, collaborator := range c.Collaborators {
		safe.Collaborators[i] = collaborator.PublicSafe()
	}
	return &safe
}

// PublicSafe returns only the identifiers of the collaborator.
func (c *Collaborator) PublicSafe() *Collaborator {
	if c == nil {
		return nil
	}
	return &Collaborator{
		OrganizationOrUserIdentifiers: c.OrganizationOrUserIdentifiers,
	}
}
