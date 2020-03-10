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
	"context"
)

type v2 struct{}

func (v2 *v2) BirthTopic(uid string) []string {
	return []string{"connect"}
}

func (v2 *v2) IsBirthTopic(path []string) bool {
	return len(path) == 1 && path[0] == "connect"
}

func (v2 *v2) LastWillTopic(uid string) []string {
	return []string{"disconnect"}
}

func (v2 *v2) IsLastWillTopic(path []string) bool {
	return len(path) == 1 && path[0] == "disconnect"
}

func (v2 *v2) UplinkTopic(uid string) []string {
	return v2.createTopic(uid, []string{"up"})
}

func (v2 *v2) IsUplinkTopic(path []string) bool {
	return len(path) == 2 && path[1] == "up"
}

func (v2 *v2) StatusTopic(uid string) []string {
	return v2.createTopic(uid, []string{"status"})
}

func (v2 *v2) IsStatusTopic(path []string) bool {
	return len(path) == 2 && path[1] == "status"
}

func (v2 *v2) TxAckTopic(uid string) []string {
	return nil
}

func (v2 *v2) IsTxAckTopic(path []string) bool {
	return false
}

func (v2 *v2) DownlinkTopic(uid string) []string {
	return v2.createTopic(uid, []string{"down"})
}

func (v2 *v2) createTopic(uid string, path []string) []string {
	inTopicIdentifier := uid
	return append([]string{inTopicIdentifier}, path...)
}

// NewV2 returns a topic layout that uses the legacy The Things Stack V2 topic structure.
func NewV2(ctx context.Context) Layout {
	return &v2{}
}
