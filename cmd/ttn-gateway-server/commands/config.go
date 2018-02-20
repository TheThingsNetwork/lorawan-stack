// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"github.com/TheThingsNetwork/ttn/cmd/internal/shared"
	shared_gatewayserver "github.com/TheThingsNetwork/ttn/cmd/internal/shared/gatewayserver"
	conf "github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver"
)

type Config struct {
	conf.ServiceBase `name:",squash"`
	GS               gatewayserver.Config `name:"gs"`
}

var DefaultConfig = Config{
	ServiceBase: shared.DefaultServiceBase,
	GS:          shared_gatewayserver.DefaultGatewayServerConfig,
}
