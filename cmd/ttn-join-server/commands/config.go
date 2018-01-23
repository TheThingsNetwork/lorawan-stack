// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"github.com/TheThingsNetwork/ttn/cmd/shared"
	conf "github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/joinserver"
)

type Config struct {
	conf.ServiceBase `name:",squash"`
	JS               joinserver.Config `name:"js"`
}

var DefaultConfig = Config{
	ServiceBase: shared.DefaultServiceBase,
	JS:          shared.DefaultJoinServerConfig,
}
