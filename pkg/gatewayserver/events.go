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

package gatewayserver

import (
	"go.thethings.network/lorawan-stack/pkg/events"
)

var (
	evtStartGatewayLink = events.Define("gateway.start_link", "start gateway link")
	evtEndGatewayLink   = events.Define("gateway.end_link", "end gateway link")

	evtReceiveUp     = events.Define("gs.up.receive", "receive uplink message")
	evtReceiveStatus = events.Define("status.receive", "receive status message")
	evtSendDown      = events.Define("gs.down.send", "send downlink message")
)
