// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"github.com/TheThingsNetwork/ttn/cmd/shared"
	conf "github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver"
)

type Config struct {
	conf.ServiceBase `name:",squash"`
	IS               identityserver.Config `name:"is"`
}

var DefaultConfig = Config{
	ServiceBase: shared.DefaultServiceBase,
	IS:          shared.DefaultIdentityServerConfig,
}
