// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package mqtt

import (
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/formatters"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/mqtt/topics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// Format represents a topic layout and message formatter.
type Format interface {
	topics.Layout
	formatters.Formatter
}

// TopicParts generates the topic parts for the provided uplink.
func TopicParts(up *io.ContextualApplicationUp, layout topics.Layout) []string {
	var f func(string, string) []string
	switch up.Up.(type) {
	case *ttnpb.ApplicationUp_UplinkMessage:
		f = layout.UplinkMessageTopic
	case *ttnpb.ApplicationUp_UplinkNormalized:
		f = layout.UplinkNormalizedTopic
	case *ttnpb.ApplicationUp_JoinAccept:
		f = layout.JoinAcceptTopic
	case *ttnpb.ApplicationUp_DownlinkAck:
		f = layout.DownlinkAckTopic
	case *ttnpb.ApplicationUp_DownlinkNack:
		f = layout.DownlinkNackTopic
	case *ttnpb.ApplicationUp_DownlinkSent:
		f = layout.DownlinkSentTopic
	case *ttnpb.ApplicationUp_DownlinkFailed:
		f = layout.DownlinkFailedTopic
	case *ttnpb.ApplicationUp_DownlinkQueued:
		f = layout.DownlinkQueuedTopic
	case *ttnpb.ApplicationUp_DownlinkQueueInvalidated:
		f = layout.DownlinkQueueInvalidatedTopic
	case *ttnpb.ApplicationUp_LocationSolved:
		f = layout.LocationSolvedTopic
	case *ttnpb.ApplicationUp_ServiceData:
		f = layout.ServiceDataTopic
	default:
		panic("unreachable")
	}
	return f(unique.ID(up.Context, up.EndDeviceIds.ApplicationIds), up.EndDeviceIds.DeviceId)
}
