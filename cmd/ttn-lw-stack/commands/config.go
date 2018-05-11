// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

// Config for the ttn-lw-stack binary.
type Config struct {
	conf.ServiceBase `name:",squash"`
	IS               identityserver.Config    `name:"is"`
	GS               gatewayserver.Config     `name:"gs"`
	NS               networkserver.Config     `name:"ns"`
	AS               applicationserver.Config `name:"as"`
	JS               joinserver.Config        `name:"js"`
}

// DefaultConfig contains the default config for the ttn-lw-stack binary.
var DefaultConfig = Config{
	ServiceBase: shared.DefaultServiceBase,
	IS:          shared_identityserver.DefaultIdentityServerConfig,
	GS:          shared_gatewayserver.DefaultGatewayServerConfig,
	NS:          shared_networkserver.DefaultNetworkServerConfig,
	AS:          shared_applicationserver.DefaultApplicationServerConfig,
	JS:          shared_joinserver.DefaultJoinServerConfig,
}
