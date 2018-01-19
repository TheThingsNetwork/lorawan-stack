// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// InvitationStore is the store that holds invitations.
type InvitationStore interface {
	// Save saves the invitation.
	Save(token, email string, expiresIn uint32) error

	// Use sets `used_at` to the current timestamp and links the invitation to an user ID.
	Use(token, userID string) error

	// Lists list all the saved invitations.
	List() ([]*ttnpb.ListInvitationsResponse_Invitation, error)

	// Delete deletes an invitation by its ID.
	Delete(id string) error
}
