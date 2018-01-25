// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email/sendgrid"
)

// Options for initializing the identity server
func Options(c *component.Component, config *identityserver.Config) (options []identityserver.Option) {
	if config.SendGridAPIKey != "" {
		options = append(options, identityserver.WithEmailProvider(sendgrid.New(c.Logger(), config.SendGridAPIKey)))
	}
	return
}
