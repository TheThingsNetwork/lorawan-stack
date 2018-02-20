// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"github.com/TheThingsNetwork/ttn/cmd/internal/shared"
	shared_applicationserver "github.com/TheThingsNetwork/ttn/cmd/internal/shared/applicationserver"
	"github.com/TheThingsNetwork/ttn/pkg/applicationserver"
	conf "github.com/TheThingsNetwork/ttn/pkg/config"
)

type Config struct {
	conf.ServiceBase `name:",squash"`
	AS               applicationserver.Config `name:"as"`
}

var DefaultConfig = Config{
	ServiceBase: shared.DefaultServiceBase,
	AS:          shared_applicationserver.DefaultApplicationServerConfig,
}
