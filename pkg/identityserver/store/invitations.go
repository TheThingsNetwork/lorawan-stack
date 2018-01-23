// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import "time"

// InvitationData is the data stored for an invitation.
type InvitationData struct {
	// Token is the secret invitation token.
	Token string

	// Email is the email the invitation was sent at.
	Email string

	// SentAt is the time when the invitation was sent.
	IssuedAt time.Time

	// ExpiresAt denotes the time the invitation will expire.
	ExpiresAt time.Time
}

// InvitationStore is the store that holds invitations.
type InvitationStore interface {
	// Save saves the invitation.
	Save(data *InvitationData) error

	// Lists list all the saved invitations.
	List() ([]*InvitationData, error)

	// Use deletes an invitation but also takes into account the token binded to it.
	Use(email, token string) error

	// Delete deletes an invitation.
	Delete(email string) error
}
