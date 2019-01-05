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

// Layout represents an MQTT topic layout.
type Layout interface {
	AcceptedTopic(applicationUID string, requested []string) (accepted []string, ok bool)

	UplinkTopic(applicationUID, deviceID string) []string
	JoinAcceptTopic(applicationUID, deviceID string) []string
	DownlinkAckTopic(applicationUID, deviceID string) []string
	DownlinkNackTopic(applicationUID, deviceID string) []string
	DownlinkSentTopic(applicationUID, deviceID string) []string
	DownlinkFailedTopic(applicationUID, deviceID string) []string
	DownlinkQueuedTopic(applicationUID, deviceID string) []string
	LocationSolvedTopic(applicationUID, deviceID string) []string

	DownlinkPushTopic(applicationUID, deviceID string) []string
	IsDownlinkPushTopic(parts []string) bool
	ParseDownlinkPushTopic(parts []string) (deviceID string)
	DownlinkReplaceTopic(applicationUID, deviceID string) []string
	IsDownlinkReplaceTopic(parts []string) bool
	ParseDownlinkReplaceTopic(parts []string) (deviceID string)
}
