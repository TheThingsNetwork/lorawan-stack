// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package redis

import (
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/events/basic"
)

type subscription struct {
	basicSub *basic.Subscription
	patterns []string
}

func (s *subscription) matchPattern(evt events.Event) bool {
	if evt, ok := evt.(*patternEvent); ok {
		for _, pattern := range s.patterns {
			if pattern == evt.pattern {
				return true
			}
		}
	}
	return false
}

func (s *subscription) Match(evt events.Event) bool {
	if s == nil {
		return false
	}
	if !s.matchPattern(evt) {
		return false
	}
	return s.basicSub.Match(evt)
}

func (s *subscription) Notify(evt events.Event) {
	if s == nil {
		return
	}
	s.basicSub.Notify(evt)
}
