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
	return m.GetEndDeviceIds().EntityType()
}

func (m *ApplicationPackageAssociation) EntityType() string {
	return m.GetIds().EntityType()
}

func (m *GetApplicationPackageAssociationRequest) EntityType() string {
	return m.GetIds().EntityType()
}

func (m *ListApplicationPackageAssociationRequest) EntityType() string {
	return m.GetIds().EntityType()
}

func (m *SetApplicationPackageAssociationRequest) EntityType() string {
	return m.GetAssociation().EntityType()
}

func (m *ApplicationPackageDefaultAssociationIdentifiers) EntityType() string {
	return m.GetApplicationIds().EntityType()
}

func (m *ApplicationPackageDefaultAssociation) EntityType() string {
	return m.GetIds().EntityType()
}

func (m *GetApplicationPackageDefaultAssociationRequest) EntityType() string {
	return m.GetIds().EntityType()
}

func (m *ListApplicationPackageDefaultAssociationRequest) EntityType() string {
	return m.GetIds().EntityType()
}

func (m *SetApplicationPackageDefaultAssociationRequest) EntityType() string {
	return m.GetDefault().EntityType()
}

// All IDString methods implement the IDStringer interface.

func (m *ApplicationPackageAssociationIdentifiers) IDString() string {
	return m.GetEndDeviceIds().IDString()
}

func (m *ApplicationPackageAssociation) IDString() string {
	return m.GetIds().IDString()
}

func (m *GetApplicationPackageAssociationRequest) IDString() string {
	return m.GetIds().IDString()
}

func (m *ListApplicationPackageAssociationRequest) IDString() string {
	return m.GetIds().IDString()
}

func (m *SetApplicationPackageAssociationRequest) IDString() string {
	return m.GetAssociation().IDString()
}

func (m *ApplicationPackageDefaultAssociationIdentifiers) IDString() string {
	return m.GetApplicationIds().IDString()
}

func (m *ApplicationPackageDefaultAssociation) IDString() string {
	return m.GetIds().IDString()
}

func (m *GetApplicationPackageDefaultAssociationRequest) IDString() string {
	return m.GetIds().IDString()
}

func (m *ListApplicationPackageDefaultAssociationRequest) IDString() string {
	return m.GetIds().IDString()
}

func (m *SetApplicationPackageDefaultAssociationRequest) IDString() string {
	return m.GetDefault().IDString()
}

// All ExtractRequestFields methods are used by github.com/grpc-ecosystem/go-grpc-middleware/tags.

func (m *ApplicationPackageAssociationIdentifiers) ExtractRequestFields(dst map[string]interface{}) {
	m.GetEndDeviceIds().ExtractRequestFields(dst)
}

func (m *ApplicationPackageAssociation) ExtractRequestFields(dst map[string]interface{}) {
	m.GetIds().ExtractRequestFields(dst)
}

func (m *GetApplicationPackageAssociationRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetIds().ExtractRequestFields(dst)
}

func (m *ListApplicationPackageAssociationRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetIds().ExtractRequestFields(dst)
}

func (m *SetApplicationPackageAssociationRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.Association.ExtractRequestFields(dst)
}

func (m *ApplicationPackageDefaultAssociationIdentifiers) ExtractRequestFields(dst map[string]interface{}) {
	m.GetApplicationIds().ExtractRequestFields(dst)
}

func (m *ApplicationPackageDefaultAssociation) ExtractRequestFields(dst map[string]interface{}) {
	m.GetIds().ExtractRequestFields(dst)
}

func (m *GetApplicationPackageDefaultAssociationRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetIds().ExtractRequestFields(dst)
}

func (m *ListApplicationPackageDefaultAssociationRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetIds().ExtractRequestFields(dst)
}

func (m *SetApplicationPackageDefaultAssociationRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetDefault().ExtractRequestFields(dst)
}
