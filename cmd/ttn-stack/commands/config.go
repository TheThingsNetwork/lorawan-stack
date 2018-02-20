// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"github.com/TheThingsNetwork/ttn/cmd/internal/shared"
	shared_applicationserver "github.com/TheThingsNetwork/ttn/cmd/internal/shared/applicationserver"
	shared_gatewayserver "github.com/TheThingsNetwork/ttn/cmd/internal/shared/gatewayserver"
	shared_identityserver "github.com/TheThingsNetwork/ttn/cmd/internal/shared/identityserver"
	shared_joinserver "github.com/TheThingsNetwork/ttn/cmd/internal/shared/joinserver"
	shared_networkserver "github.com/TheThingsNetwork/ttn/cmd/internal/shared/networkserver"
	"github.com/TheThingsNetwork/ttn/pkg/applicationserver"
	conf "github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver"
	"github.com/TheThingsNetwork/ttn/pkg/joinserver"
	"github.com/TheThingsNetwork/ttn/pkg/networkserver"
)

type Config struct {
	conf.ServiceBase `name:",squash"`
	IS               identityserver.Config    `name:"is"`
	GS               gatewayserver.Config     `name:"gs"`
	NS               networkserver.Config     `name:"ns"`
	AS               applicationserver.Config `name:"as"`
	JS               joinserver.Config        `name:"js"`
}

var DefaultConfig = Config{
	ServiceBase: shared.DefaultServiceBase,
	IS:          shared_identityserver.DefaultIdentityServerConfig,
	GS:          shared_gatewayserver.DefaultGatewayServerConfig,
	NS:          shared_networkserver.DefaultNetworkServerConfig,
	AS:          shared_applicationserver.DefaultApplicationServerConfig,
	JS:          shared_joinserver.DefaultJoinServerConfig,
}
