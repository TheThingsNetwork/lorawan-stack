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

	evts := events.Builders(builders).New(ctx)
	log.FromContext(ctx).WithField("event_count", n).Debug("Publish events")
	events.Publish(evts...)
}

func eventOptions(extraOpts ...events.Option) []events.Option {
	return append([]events.Option{events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ)}, extraOpts...)
}

func defineCmdReceivedEvent(name, desc string, opts ...events.Option) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("as.packages.alcsync.v1.%s.cmd_received", name),
		fmt.Sprintf("%s command received", desc),
		eventOptions(opts...)...,
	)
}

func defineCmdAnsEnqueueEvent(name, desc string, opts ...events.Option) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("as.packages.alcsync.v1.%s.answer_enqueued", name),
		fmt.Sprintf("%s command answer enqueued", desc),
		eventOptions(opts...)...,
	)
}

var (
	// EvtTimeCorrectionCmdReceived is the event that is published when
	// a time correction command is received and successfully parsed.
	EvtTimeCorrectionCmdReceived = defineCmdReceivedEvent(
		"time_correction", "time correction",
		events.WithDataType(&ttnpb.ALCSyncCommand_AppTimeReq{}),
	)()

	// EvtTimeCorrectionAnsEnqueue is the event that is published when
	// a time correction command is processed successfully and an answer is enqueued.
	EvtTimeCorrectionAnsEnqueue = defineCmdAnsEnqueueEvent(
		"time_correction", "time correction",
		events.WithDataType(&ttnpb.ALCSyncCommand_AppTimeAns{}),
	)()

	// EvtPkgFail is the event that is published when an error occurs in the package.
	EvtPkgFail = events.Define(
		"as.packages.alcsync.v1.fail", "package failed due to error", eventOptions(
			events.WithErrorDataType(), events.WithPropagateToParent(),
		)...,
	)
)
