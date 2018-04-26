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

package identityserver

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// authorizationData is the type that abstracts the authorization data contained in a request.
type authorizationData struct {
	// EntityIdentifiers contains the ttnpb.XXXIdentifiers of the entity that the
	// authorization data is related to.
	//
	// If the data comes from an access token then it will be a ttnpb.UserIdentifiers,
	// otherwise, if it comes from an API key it will be a ttnpb.XXXIdentifiers
	// that matches the API key type, e.g. a ttnpb.ApplicationIdentifiers for
	// an application API key.
	EntityIdentifiers interface{}
	// Source denotes if the data comes from an API key or Access Token.
	Source string
	// Rights are either the API key rights or the Access Token scope.
	Rights []ttnpb.Right
}

// UserIdentifiers returns the ttnpb.UserIdentifiers of the user the authorization
// data is related to, otherwise, a zero-valued ttnpb.UserIdentifiers.
func (ad *authorizationData) UserIdentifiers() (ids ttnpb.UserIdentifiers) {
	if i, ok := ad.EntityIdentifiers.(ttnpb.UserIdentifiers); ok {
		ids = i
	}
	return
}

// ApplicationIdentifiers returns the ttnpb.ApplicationIdentifiers of the application
// the authorization data is related to, otherwise, a zero-valued ttnpb.ApplicationIdentifiers.
func (ad *authorizationData) ApplicationIdentifiers() (ids ttnpb.ApplicationIdentifiers) {
	if i, ok := ad.EntityIdentifiers.(ttnpb.ApplicationIdentifiers); ok {
		ids = i
	}
	return
}

// GatewayIdentifiers returns the ttnpb.GatewayIdentifiers of the gateway the
// authorization data is related to, otherwise, a zero-valued ttnpb.GatewayIdentifiers.
func (ad *authorizationData) GatewayIdentifiers() (ids ttnpb.GatewayIdentifiers) {
	if i, ok := ad.EntityIdentifiers.(ttnpb.GatewayIdentifiers); ok {
		ids = i
	}
	return
}

// OrganizationIdentifiers returns the ttnpb.OrganizationIdentifiers of the organization the
// authorization data is related to, otherwise, a zero-valued ttnpb.OrganizationIdentifiers.
func (ad *authorizationData) OrganizationIdentifiers() (ids ttnpb.OrganizationIdentifiers) {
	if i, ok := ad.EntityIdentifiers.(ttnpb.OrganizationIdentifiers); ok {
		ids = i
	}
	return
}

// HasRights checks whether or not the provided rights are included in the authorization data.
// It will only return true if all the provided rights are included in the authorization data.
func (ad *authorizationData) HasRights(rights ...ttnpb.Right) bool {
	ok := true
	for _, right := range rights {
		ok = ok && ad.hasRight(right)
	}

	return ok
}

// hasRight checks whether or not a right is included in this authorization data.
func (ad *authorizationData) hasRight(right ttnpb.Right) bool {
	for _, r := range ad.Rights {
		if r == right {
			return true
		}
	}
	return false
}
