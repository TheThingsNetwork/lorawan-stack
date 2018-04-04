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

package store

// Store is a store that holds all different sub-stores.
type Store struct {
	// Users is the users store.
	Users UserStore

	// Applications is the applications store.
	Applications ApplicationStore

	// Gateways is the gateways store.
	Gateways GatewayStore

	// Clients is the clients store.
	Clients ClientStore

	// OAuth is the OAuth store.
	OAuth OAuthStore

	// Settings is the settings store.
	Settings SettingStore

	// Invitations is the invitations store.
	Invitations InvitationStore

	// Organizations is the organizations store.
	Organizations OrganizationStore
}

// Attributer is the interface providing methods to extend basic IS data types.
//
// Types implementing the Attributer interface are able to extend its
// default data type by having extra attributes that are stored in different
// tables or collection of the data store.
type Attributer interface {
	// Namespaces returns all namespaces the type can have extra attributes in.
	// This is useful for splitting up the attributes into multiple smaller
	// types that all wrap each other.
	Namespaces() []string

	// Fill fills the type extra attributes from the key values that were found
	// in the store.
	Fill(namespace string, attributes map[string]interface{}) error

	// Attributes returns all extra attributes of the type as a key value map.
	// These attributes will be stored in the store on create or update.
	Attributes(namespace string) map[string]interface{}
}
