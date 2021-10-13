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

func (m *ApplicationPackageAssociationIdentifiers) EntityType() string {
	return m.EndDeviceIds.EntityType()
}

func (m *ApplicationPackageAssociation) EntityType() string {
	return m.Ids.EntityType()
}

func (m *GetApplicationPackageAssociationRequest) EntityType() string {
	return m.Ids.EntityType()
}

func (m *ListApplicationPackageAssociationRequest) EntityType() string {
	return m.Ids.EntityType()
}

func (m *SetApplicationPackageAssociationRequest) EntityType() string {
	return m.Association.Ids.EntityType()
}

func (m *ApplicationPackageDefaultAssociationIdentifiers) EntityType() string {
	return m.ApplicationIds.EntityType()
}

func (m *ApplicationPackageDefaultAssociation) EntityType() string {
	return m.Ids.EntityType()
}

func (m *GetApplicationPackageDefaultAssociationRequest) EntityType() string {
	return m.Ids.EntityType()
}

func (m *ListApplicationPackageDefaultAssociationRequest) EntityType() string {
	return m.Ids.EntityType()
}

func (m *SetApplicationPackageDefaultAssociationRequest) EntityType() string {
	return m.Default.Ids.EntityType()
}

// All IDString methods implement the IDStringer interface.

func (m *ApplicationPackageAssociationIdentifiers) IDString() string {
	return m.EndDeviceIds.IDString()
}

func (m *ApplicationPackageAssociation) IDString() string {
	return m.Ids.IDString()
}

func (m *GetApplicationPackageAssociationRequest) IDString() string {
	return m.Ids.IDString()
}

func (m *ListApplicationPackageAssociationRequest) IDString() string {
	return m.Ids.IDString()
}

func (m *SetApplicationPackageAssociationRequest) IDString() string {
	return m.Association.Ids.IDString()
}

func (m *ApplicationPackageDefaultAssociationIdentifiers) IDString() string {
	return m.ApplicationIds.IDString()
}

func (m *ApplicationPackageDefaultAssociation) IDString() string {
	return m.Ids.IDString()
}

func (m *GetApplicationPackageDefaultAssociationRequest) IDString() string {
	return m.Ids.IDString()
}

func (m *ListApplicationPackageDefaultAssociationRequest) IDString() string {
	return m.Ids.IDString()
}

func (m *SetApplicationPackageDefaultAssociationRequest) IDString() string {
	return m.Default.Ids.IDString()
}

// All ExtractRequestFields methods are used by github.com/grpc-ecosystem/go-grpc-middleware/tags.

func (m *ApplicationPackageAssociationIdentifiers) ExtractRequestFields(dst map[string]interface{}) {
	m.EndDeviceIds.ExtractRequestFields(dst)
}

func (m *ApplicationPackageAssociation) ExtractRequestFields(dst map[string]interface{}) {
	m.Ids.ExtractRequestFields(dst)
}

func (m *GetApplicationPackageAssociationRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.Ids.ExtractRequestFields(dst)
}

func (m *ListApplicationPackageAssociationRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.Ids.ExtractRequestFields(dst)
}

func (m *SetApplicationPackageAssociationRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.Association.Ids.ExtractRequestFields(dst)
}

func (m *ApplicationPackageDefaultAssociationIdentifiers) ExtractRequestFields(dst map[string]interface{}) {
	m.ApplicationIds.ExtractRequestFields(dst)
}

func (m *ApplicationPackageDefaultAssociation) ExtractRequestFields(dst map[string]interface{}) {
	m.Ids.ExtractRequestFields(dst)
}

func (m *GetApplicationPackageDefaultAssociationRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.Ids.ExtractRequestFields(dst)
}

func (m *ListApplicationPackageDefaultAssociationRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.Ids.ExtractRequestFields(dst)
}

func (m *SetApplicationPackageDefaultAssociationRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.Default.Ids.ExtractRequestFields(dst)
}
