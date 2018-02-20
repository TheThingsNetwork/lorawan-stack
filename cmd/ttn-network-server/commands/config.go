// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"github.com/TheThingsNetwork/ttn/cmd/internal/shared"
	shared_networkserver "github.com/TheThingsNetwork/ttn/cmd/internal/shared/networkserver"
	conf "github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/networkserver"
)

type Config struct {
	conf.ServiceBase `name:",squash"`
	NS               networkserver.Config `name:"ns"`
}

var DefaultConfig = Config{
	ServiceBase: shared.DefaultServiceBase,
	NS:          shared_networkserver.DefaultNetworkServerConfig,
}
