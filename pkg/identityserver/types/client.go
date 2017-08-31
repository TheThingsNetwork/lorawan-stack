// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import "time"

// ClientState represents the review state of a client by the staff
type ClientState int

const (
	// PendingClient means that the client has not been reviewed by the staff yet
	PendingClient ClientState = iota

	// ApprovedClient means that the client has been approved and thus can be used
	ApprovedClient

	// RejectedClient means that the client has been rejected to be used
	RejectedClient
)

// DefaultClient represents a third-party client
type DefaultClient struct {
	// ID is the unique client identifier
	ID string `db:"id"`

	// Description is the description of the client
	Description *string `db:"description"`

	// Secret is the secret used to prove the client identity
	Secret string `db:"secret"`

	// URI is the callback URI of the client
	URI string `db:"uri"`

	// State denotes the reviewing state by the staff of the client
	State ClientState `db:"state"`

	// Official denotes if the client is a client created by the staff
	Official bool `db:"official"`

	// Grants denotes which OAuth2 flows can the client use to get a token
	Grants Grants `db:"grants"`

	// Scope denotes what scopes the client will have access to
	Scope Scopes `db:"scope"`

	// Created denotes when the client was created
	Created time.Time `db:"created"`

	// Archived denotes when the client was archived and therefore disabled
	Archived *time.Time `db:"archived"`
}

// Client is the interface that represents a Client
type Client interface {
	// GetClient returns the base
	GetClient() *DefaultClient
}

// GetClient implements Client interface
func (d *DefaultClient) GetClient() *DefaultClient {
	return d
}

// String implements fmt.Stringer interface
func (s ClientState) String() string {
	switch s {
	case PendingClient:
		return "Pending"
	case ApprovedClient:
		return "Approved"
	case RejectedClient:
		return "Rejected"
	default:
		return "Invalid state"
	}
}
