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

package topics

import (
	"github.com/TheThingsIndustries/mystique/pkg/topic"
)

const topicV3 = "v3"

type v3 struct{}

func (v3) AcceptedTopic(applicationUID string, requested []string) ([]string, bool) {
	// Rewrite # to v3/uid/#
	if requested[0] == topic.Wildcard {
		return []string{topicV3, applicationUID, topic.Wildcard}, true
	}
	if requested[0] != topicV3 || len(requested) < 2 {
		return nil, false
	}
	switch requested[1] {
	case topic.Wildcard:
		// Rewrite v3/# to v3/uid/#
		return []string{topicV3, applicationUID, topic.Wildcard}, true
	case topic.PartWildcard:
		// Rewrite v3/+/... to v3/uid/...
		requested[1] = applicationUID
		return requested, true
	case applicationUID:
		return requested, true
	}
	return nil, false
}

func (v3) UplinkTopic(applicationUID, deviceID string) []string {
	return []string{topicV3, applicationUID, "devices", deviceID, "up"}
}

func (v3) JoinAcceptTopic(applicationUID, deviceID string) []string {
	return []string{topicV3, applicationUID, "devices", deviceID, "join"}
}

func (v3) DownlinkAckTopic(applicationUID, deviceID string) []string {
	return []string{topicV3, applicationUID, "devices", deviceID, "down", "ack"}
}

func (v3) DownlinkNackTopic(applicationUID, deviceID string) []string {
	return []string{topicV3, applicationUID, "devices", deviceID, "down", "nack"}
}

func (v3) DownlinkSentTopic(applicationUID, deviceID string) []string {
	return []string{topicV3, applicationUID, "devices", deviceID, "down", "sent"}
}

func (v3) DownlinkFailedTopic(applicationUID, deviceID string) []string {
	return []string{topicV3, applicationUID, "devices", deviceID, "down", "failed"}
}

func (v3) DownlinkQueuedTopic(applicationUID, deviceID string) []string {
	return []string{topicV3, applicationUID, "devices", deviceID, "down", "queued"}
}

func (v3) LocationSolvedTopic(applicationUID, deviceID string) []string {
	return []string{topicV3, applicationUID, "devices", deviceID, "location", "solved"}
}

func (v3) DownlinkPushTopic(applicationUID, deviceID string) []string {
	return []string{topicV3, applicationUID, "devices", deviceID, "down", "push"}
}

func (v3) IsDownlinkPushTopic(parts []string) bool {
	return len(parts) == 6 && parts[0] == topicV3 && parts[2] == "devices" && parts[4] == "down" && parts[5] == "push"
}

func (v3) ParseDownlinkPushTopic(parts []string) (deviceID string) {
	return parts[3]
}

func (v3) DownlinkReplaceTopic(applicationUID, deviceID string) []string {
	return []string{topicV3, applicationUID, "devices", deviceID, "down", "replace"}
}

func (v3) IsDownlinkReplaceTopic(parts []string) bool {
	return len(parts) == 6 && parts[0] == topicV3 && parts[2] == "devices" && parts[4] == "down" && parts[5] == "replace"
}

func (v3) ParseDownlinkReplaceTopic(parts []string) (deviceID string) {
	return parts[3]
}

// Default is the default topic layout.
var Default Layout = &v3{}
