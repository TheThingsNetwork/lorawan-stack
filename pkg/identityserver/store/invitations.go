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
	Save(data InvitationData) error

	// Lists list all the saved invitations.
	List() ([]*InvitationData, error)

	// Use deletes an invitation but also takes into account the token binded to it.
	Use(email, token string) error

	// Delete deletes an invitation.
	Delete(email string) error
}
