// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package claims

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// Entity is the enum that defines the entity types the claims are intended for.
type Entity int

const (
	// User represents a claims for an user.
	User Entity = iota

	// Application represents a claims for an application.
	Application

	// Gateway represents a claims for a gateway.
	Gateway

	// Organization represents a claims for an organization.
	Organization
)

// Claims is the type that represents a claims to do something in the Identity Server.
type Claims struct {
	// entityID is the ID of the entity this claims are intended for.
	entityID string

	// entityType is the type of entity this claims are intended for.
	entityType Entity

	// Source is the source of this claims, either an API key or a token.
	source string

	// Rights is the list of actions this token or API key has access to.
	rights []ttnpb.Right
}

// New returns a filled Claims.
func New(id string, typ Entity, source string, rights []ttnpb.Right) *Claims {
	return &Claims{
		entityID:   id,
		entityType: typ,
		source:     source,
		rights:     rights,
	}
}

// UserID returns the user ID of the user profile this claims are for, or the
// empty string if it is not for a user.
func (c *Claims) UserID() (id string) {
	if c.entityType == User {
		id = c.entityID
	}
	return
}

// ApplicationID returns the application ID  of the application this claims are
// for, or the empty string if it is not for an application.
func (c *Claims) ApplicationID() (id string) {
	if c.entityType == Application {
		id = c.entityID
	}
	return
}

// GatewayID returns the gateway ID of the gateway this claims are for, or the
// empty string if it is not for a gateway.
func (c *Claims) GatewayID() (id string) {
	if c.entityType == Gateway {
		id = c.entityID
	}
	return
}

// OrganizationID returns the organization ID of the organization this claims are
// for, or the empty string if it is not for an organization.
func (c *Claims) OrganizationID() (id string) {
	if c.entityType == Organization {
		id = c.entityID
	}
	return
}

// Source returns the value's type the claims were formed, either an API key
// or an access token.
func (c *Claims) Source() string {
	return c.source
}

// Rights returns the list of rights the caller has access to the entity.
func (c *Claims) Rights() []ttnpb.Right {
	return c.rights
}

// HasRights checks whether or not the provided rights are included in the claims.
// It will only return true if all the provided rights are included in the claims.
func (c *Claims) HasRights(rights ...ttnpb.Right) bool {
	ok := true
	for _, right := range rights {
		ok = ok && c.hasRight(right)
	}

	return ok
}

// hasRight checks whether or not the right is included in this claims.
func (c *Claims) hasRight(right ttnpb.Right) bool {
	for _, r := range c.rights {
		if r == right {
			return true
		}
	}
	return false
}
