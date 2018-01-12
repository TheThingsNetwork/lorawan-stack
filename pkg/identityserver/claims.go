// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// entity is an enum that defines the valids entity types for claims.
type entity int

const (
	// EntityUser is an user.
	entityUser = iota

	// EntityApplication is an application.
	entityApplication

	// EntityGateway is a gateway.
	entityGateway
)

// claims is the type that represents a claims to do something in the Identity Server.
type claims struct {
	// EntityID is the ID of the entity this claims are intended for.
	EntityID string

	// EntityType is the type of entity this claims are intended for.
	EntityType entity

	// Source is the source of this claims, either an API key or a token.
	Source string

	// Rights is the list of actions this token has access to.
	Rights []ttnpb.Right
}

// UserID returns the user ID of the user profile this claims are for, or the
// empty string if it is not for a user.
func (c *claims) UserID() (id string) {
	if c.EntityType == entityUser {
		id = c.EntityID
	}
	return
}

// ApplicationID returns the application ID  of the application this claims are
// for, or the empty string if it is not for an application.
func (c *claims) ApplicationID() (id string) {
	if c.EntityType == entityApplication {
		id = c.EntityID
	}
	return
}

// GatewayID returns the gateway ID of the gateway this claims are for, or the
// empty string if it is not for a gateway.
func (c *claims) GatewayID() (id string) {
	if c.EntityType == entityGateway {
		id = c.EntityID
	}
	return
}

// HasRights checks whether or not the provided rights are included in the claims.
// It will only return true if all the provided rights are included in the claims.
func (c *claims) HasRights(rights ...ttnpb.Right) bool {
	ok := true
	for _, right := range rights {
		ok = ok && c.hasRight(right)
	}

	return ok
}

// hasRight checks whether or not the right is included in this claims.
func (c *claims) hasRight(right ttnpb.Right) bool {
	for _, r := range c.Rights {
		if r == right {
			return true
		}
	}
	return false
}
