// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package lbslns

import (
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws"
)

var (
	errSessionStateNotFound = errors.DefineUnavailable("session_state_not_found", "session state not found")
	trafficEndPointPrefix   = "/traffic"
)

// State represents the LBS Session state.
type State struct {
	ID int32
}

type lbsLNS struct {
	tokens io.DownlinkTokens
}

// NewFormatter returns a new LoRa Basic Station LNS formatter.
func NewFormatter() ws.Formatter {
	return &lbsLNS{}
}

func (f *lbsLNS) Endpoints() ws.Endpoints {
	return ws.Endpoints{
		ConnectionInfo: "/router-info",
		Traffic:        fmt.Sprintf("%s/:id", trafficEndPointPrefix),
	}
}
