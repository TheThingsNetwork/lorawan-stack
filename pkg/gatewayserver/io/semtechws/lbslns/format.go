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
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/semtechws"
)

var (
	errSessionStateNotFound = errors.DefineUnavailable("session_state_not_found", "session state not found")
	trafficEndPointPrefix   = "/traffic"
)

type lbsLNS struct {
	maxRoundTripDelay time.Duration
}

// NewFormatter returns a new LoRa Basic Station LNS formatter.
func NewFormatter(maxRoundTripDelay time.Duration) semtechws.Formatter {
	return &lbsLNS{
		maxRoundTripDelay: maxRoundTripDelay,
	}
}

func (f *lbsLNS) ID() string { return "lbslns" }

func (f *lbsLNS) Endpoints() semtechws.Endpoints {
	return semtechws.Endpoints{
		ConnectionInfo: "/router-info",
		Traffic:        fmt.Sprintf("%s/{id}", trafficEndPointPrefix),
	}
}
