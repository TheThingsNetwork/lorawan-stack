// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package email

import "github.com/TheThingsNetwork/ttn/pkg/identityserver/email/templates"

// Provider is the interface that describes all the email providers that can
// be used by the Identity Server.
type Provider interface {
	// Send sends an email to recipient using a template compiled with the
	// provided data.
	Send(recipient string, tmpl *templates.Template, data interface{}) error
}
