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

const (
	// FieldPathOrganizationName is the field path for the organization name field.
	FieldPathOrganizationName = "name"

	// FieldPathOrganizationDescription is the field path for the organization description field.
	FieldPathOrganizationDescription = "description"

	// FieldPathOrganizationURL is the field path for the organization URL field.
	FieldPathOrganizationURL = "url"

	// FieldPathOrganizationLocation is the field path for the organization location field.
	FieldPathOrganizationLocation = "location"

	// FieldPathOrganizationEmail is the field path for the organization email field.
	FieldPathOrganizationEmail = "email"
)

// GetOrganization returns the base Organization itself.
func (d *Organization) GetOrganization() *Organization {
	return d
}
