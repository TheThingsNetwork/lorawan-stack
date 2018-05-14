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

package networkserver

import "go.thethings.network/lorawan-stack/pkg/events"

var (
	evtStartApplicationLink = events.Define("application.start_link", "start application link")
	evtEndApplicationLink   = events.Define("application.end_link", "end application link")

	evtReceiveUp          = events.Define("up.receive", "receive uplink message")
	evtReceiveUpDuplicate = events.Define("up.receive_duplicate", "receive duplicate uplink message")
	evtMergeMetadata      = events.Define("up.merge_metadata", "merge uplink message metadata")

	evtDropData    = events.Define("up.data.drop", "drop data message")
	evtForwardData = events.Define("up.data.forward", "forward data message")

	evtDropJoin    = events.Define("up.join.drop", "drop join request")
	evtForwardJoin = events.Define("up.join.forward", "forward join request")

	evtDropRejoin    = events.Define("up.rejoin.drop", "drop rejoin request")
	evtForwardRejoin = events.Define("up.rejoin.forward", "forward rejoin request")
)
