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

// CombinedIdentifiers returns the ApplicationIdentifiers as CombinedIdentifiers.
func (ids ApplicationIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return ids.EntityIdentifiers().CombinedIdentifiers()
}

// CombinedIdentifiers returns the ClientIdentifiers as CombinedIdentifiers.
func (ids ClientIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return ids.EntityIdentifiers().CombinedIdentifiers()
}

// CombinedIdentifiers returns the EndDeviceIdentifiers as CombinedIdentifiers.
func (ids EndDeviceIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return ids.EntityIdentifiers().CombinedIdentifiers()
}

// CombinedIdentifiers returns the GatewayIdentifiers as CombinedIdentifiers.
func (ids GatewayIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return ids.EntityIdentifiers().CombinedIdentifiers()
}

// CombinedIdentifiers returns the OrganizationIdentifiers as CombinedIdentifiers.
func (ids OrganizationIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return ids.EntityIdentifiers().CombinedIdentifiers()
}

// CombinedIdentifiers returns the UserIdentifiers as CombinedIdentifiers.
func (ids UserIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return ids.EntityIdentifiers().CombinedIdentifiers()
}

// CombinedIdentifiers returns the EntityIdentifiers as a CombinedIdentifiers type.
func (ids EntityIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return &CombinedIdentifiers{EntityIdentifiers: []*EntityIdentifiers{&ids}}
}

// CombinedIdentifiers returns the OrganizationOrUserIdentifiers as CombinedIdentifiers.
func (ids OrganizationOrUserIdentifiers) CombinedIdentifiers() *CombinedIdentifiers {
	return ids.EntityIdentifiers().CombinedIdentifiers()
}
