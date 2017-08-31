// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

// Store is a store that holds all different sub-stores
type Store struct {
	// Users is the users store
	Users UserStore

	// Applications is the applications store
	Applications ApplicationStore

	// Gateways is the gateways store
	Gateways GatewayStore

	// Components is the components store
	Components ComponentStore

	// Clients is the clients store
	Clients ClientStore
}

// Attributer is the interface providing methods to extend basic IS data types
//
// Types implementing the Attributer interface are able to extend its
// default data type by having extra attributes that are stored in different
// tables or collection of the data store
type Attributer interface {
	// Namespaces returns all namespaces the type can have extra attributes in.
	// This is useful for splitting up the attributes into multiple smaller
	// types that all wrap each other.
	Namespaces() []string

	// Fill fills the type extra attributes from the key values that were found
	// in the store
	Fill(namespace string, attributes map[string]interface{}) error

	// Attributes returns all extra attributes of the type as a key value map.
	// These attributes will be stored in the store on create or update.
	Attributes(namespace string) map[string]interface{}
}
