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

package topics

const (
	topicVersion  = "v3"
	topicUplink   = "up"
	topicDownlink = "down"
	topicStatus   = "status"
)

// Uplink returns the uplink topic path.
func Uplink(uid string) []string {
	return []string{topicVersion, uid, topicUplink}
}

// IsUplink returns whether the topic is an uplink topic.
func IsUplink(topicPath []string) bool {
	if len(topicPath) != 3 {
		return false
	}
	return topicPath[0] == topicVersion && topicPath[2] == topicUplink
}

// Downlink returns the downlink topic path.
func Downlink(uid string) []string {
	return []string{topicVersion, uid, topicDownlink}
}

// IsDownlink returns whether the topic is a downlink topic.
func IsDownlink(topicPath []string) bool {
	if len(topicPath) != 3 {
		return false
	}
	return topicPath[0] == topicVersion && topicPath[2] == topicDownlink
}

// Status returns the status topic path.
func Status(uid string) []string {
	return []string{topicVersion, uid, topicStatus}
}

// IsStatus returns whether the topic is a status topic.
func IsStatus(topicPath []string) bool {
	if len(topicPath) != 3 {
		return false
	}
	return topicPath[0] == topicVersion && topicPath[2] == topicStatus
}
