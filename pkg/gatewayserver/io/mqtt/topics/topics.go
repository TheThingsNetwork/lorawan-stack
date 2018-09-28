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

// Version is the version of a message.
type Version string

const (
	topicUplink   = "up"
	topicDownlink = "down"
	topicStatus   = "status"
	topicAck      = "ack"
)

var (
	// V3 represents a v3-compatible message.
	V3 Version = "v3"
	// V2 represents a v2-compatible message.
	V2 Version = "v2"
)

func isMessage(topicPath []string, message string) bool {
	if len(topicPath) > 0 && Version(topicPath[0]) == V3 {
		topicPath = topicPath[1:]
	}
	return len(topicPath) == 2 && topicPath[1] == message
}

func makeTopicPath(uid, message string, v Version) []string {
	if v == V2 {
		return []string{uid, message}
	}
	return []string{string(v), uid, message}
}

// Uplink returns the uplink topic path.
func Uplink(uid string, v Version) []string {
	return makeTopicPath(uid, topicUplink, v)
}

// IsUplink returns whether the topic is an uplink topic.
func IsUplink(topicPath []string) bool {
	return isMessage(topicPath, topicUplink)
}

// Downlink returns the downlink topic path.
func Downlink(uid string, v Version) []string {
	return makeTopicPath(uid, topicDownlink, v)
}

// IsDownlink returns whether the topic is a downlink topic.
func IsDownlink(topicPath []string) bool {
	return isMessage(topicPath, topicDownlink)
}

// Status returns the status topic path.
func Status(uid string, v Version) []string {
	return makeTopicPath(uid, topicStatus, v)
}

// IsStatus returns whether the topic is a status topic.
func IsStatus(topicPath []string) bool {
	return isMessage(topicPath, topicStatus)
}

// TxAck returns the ack topic path.
func TxAck(uid string, v Version) []string {
	return append(Downlink(uid, v), topicAck)
}

// IsTxAck returns whether the topic is a Tx acknowledgment topic.
func IsTxAck(topicPath []string) bool {
	return len(topicPath) > 2 && isMessage(topicPath[:len(topicPath)-1], topicDownlink) && topicPath[len(topicPath)-1] == topicAck
}
