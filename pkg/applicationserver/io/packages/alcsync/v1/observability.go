// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package alcsyncv1

import (
	"context"
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func publishEvents(ctx context.Context, builders ...events.Builder) {
	n := len(builders)
	if n == 0 {
		return
	}

	evts := make([]events.Event, n)
	for i, builder := range builders {
		evts[i] = builder.New(ctx)
	}
	log.FromContext(ctx).WithField("event_count", n).Debug("Publish events")
	events.Publish(evts...)
}

func alcsyncEventOptions(extraOpts ...events.Option) []events.Option {
	return append([]events.Option{events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ)}, extraOpts...)
}

func defineALCSyncPkgFailEvent(name, desc string, opts ...events.Option) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("as.packages.alcsync.%s.failed", name), fmt.Sprintf("%s package failed", desc),
		alcsyncEventOptions(opts...)...,
	)
}

func defineALCSyncCmdReceivedEvent(name, desc string, opts ...events.Option) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("as.packages.alcsync.cmd.%s.accept", name), fmt.Sprintf("%s command received", desc),
		alcsyncEventOptions(opts...)...,
	)
}

func defineALCSyncCmdParsedEvent(name, desc string, opts ...events.Option) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("as.packages.alcsync.cmd.%s.parsed", name), fmt.Sprintf("%s command parsed", desc),
		alcsyncEventOptions(opts...)...,
	)
}

func defineALCSyncCmdParsedFailEvent(name, desc string, opts ...events.Option) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("as.packages.alcsync.cmd.%s.parse.failed", name), fmt.Sprintf("%s command failed when parsing", desc),
		alcsyncEventOptions(opts...)...,
	)
}

func defineALCSyncCmdHandledEvent(name, desc string, opts ...events.Option) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("as.packages.alcsync.cmd.%s.handled", name), fmt.Sprintf("%s command handled", desc),
		alcsyncEventOptions(opts...)...,
	)
}

func defineALCSyncCmdHandledFailEvent(name, desc string, opts ...events.Option) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("as.packages.alcsync.cmd.%s.handled.failed", name), fmt.Sprintf("%s command handled", desc),
		alcsyncEventOptions(opts...)...,
	)
}

// General package events.
var (
	// EvtALCSyncCmdReceived is the event that is published when a alcsync command is received.
	EvtALCSyncCmdReceived = defineALCSyncCmdReceivedEvent("alcsync", "alcsync package")()

	// EvtALCSyncCmdParsedFail is the event that is published when a time synchronization command fails to be parsed.
	EvtALCSyncCmdParsedFail = defineALCSyncCmdParsedFailEvent(
		"alcsync", "alcsync package", events.WithErrorDataType(), events.WithPropagateToParent(),
	)()

	// EvtALCSyncCmdHandled is the event that is published when the alsync package fails to handle a command.
	EvtALCSyncPkgFail = defineALCSyncPkgFailEvent(
		"alcsync", "alcsync", events.WithErrorDataType(), events.WithPropagateToParent(),
	)()
)

// Events for time synchronization command.
var (
	// EvtTimeSyncCmdParsed is the event that is published when a time synchronization command is successfully parsed.
	EvtTimeSyncCmdParsed = defineALCSyncCmdParsedEvent("time_sync", "Time Synchonization")()

	// EvtALCSyncCmdParsedFail is the event that is published when a time synchronization command fails to be parsed.
	EvtTimeSyncCmdParsedFali = defineALCSyncCmdParsedFailEvent(
		"time_sync", "Time Synchonization", events.WithErrorDataType(), events.WithPropagateToParent(),
	)()

	// EvtTimeSyncCmdHandled is the event that is published when a time synchronization command is successfully handled.
	EvtTimeSyncCmdHandled = defineALCSyncCmdHandledEvent("time_sync", "Time Synchonization")()

	// EvtTimeSyncCmdHandledFail is the event that is published when a time synchronization command fails to be handled.
	EvtTimeSyncCmdHandledFail = defineALCSyncCmdHandledFailEvent(
		"time_sync", "Time Synchonization", events.WithErrorDataType(), events.WithPropagateToParent(),
	)()
)
