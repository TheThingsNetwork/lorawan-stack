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

package emails

import "go.thethings.network/lorawan-stack/pkg/ttnpb"

// Data for emails.
// Update doc/content/reference/email-templates/available.md when changing fields or adding new emails.
type Data struct {
	// User we're sending this email to. We need at least an Email.
	User struct {
		ID    string
		Name  string
		Email string
	}

	// Network information to fill into the template.
	Network struct {
		Name              string
		IdentityServerURL string
		ConsoleURL        string
	}

	// Entity this is concerning
	Entity struct {
		Type string
		ID   string
	}

	// Contact details used to inform the user why they are receiving an email.
	// For example:
	//     You are receiving this because you are {{.Contact.Type}} contact on {{.Entity.Type}} {{.Entity.ID}}.
	Contact struct {
		Type string // contact type: technical, billing, abuse; see *ttnpb.ContactInfo
	}
}

// SetUser sets the user's ID, name and primary email address to the email data.
// If the user's name is unknown, its ID is used as name.
func (d *Data) SetUser(user *ttnpb.User) {
	d.User.ID = user.UserID
	if user.Name != "" {
		d.User.Name = user.Name
	} else {
		d.User.Name = user.UserID
	}
	d.User.Email = user.PrimaryEmailAddress
}

// SetEntity sets the entity that the email is about.
func (d *Data) SetEntity(ids *ttnpb.EntityIdentifiers) {
	d.Entity.Type = ids.EntityType()
	d.Entity.ID = ids.IDString()
}

// SetContact sets the contact info as recipient of the email.
func (d *Data) SetContact(contact *ttnpb.ContactInfo) {
	d.Contact.Type = contact.ContactType.String()
	if contact.ContactMethod == ttnpb.CONTACT_METHOD_EMAIL {
		d.User.Email = contact.Value
	}
	if d.User.Name == "" {
		d.User.Name = "user"
	}
}

// Recipient returns the recipient info of the email.
func (d Data) Recipient() (name, address string) {
	return d.User.Name, d.User.Email
}
