// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/cmd/internal/shared"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver"
)

// DefaultIdentityServerConfig is the default configuration for the IdentityServer.
var DefaultIdentityServerConfig = identityserver.Config{
	DatabaseURI:      "postgres://root@localhost:26257/is_development?sslmode=disable",
	PublicURL:        fmt.Sprintf("http://localhost:%s", shared.DefaultServiceBase.HTTP.Listen),
	OrganizationName: "The Things Network",
	DefaultSettings:  identityserver.DefaultSettings,
	Specializers:     identityserver.DefaultSpecializers,
}
