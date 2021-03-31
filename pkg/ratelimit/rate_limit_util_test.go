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

package ratelimit_test

import "go.thethings.network/lorawan-stack/v3/pkg/ratelimit"

var (
	maxRate      uint = 10
	overrideRate uint = 15
)

type mockLimiter struct {
	calledWithResource ratelimit.Resource

	limit  bool
	result ratelimit.Result
}

func (l *mockLimiter) RateLimit(resource ratelimit.Resource) (bool, ratelimit.Result) {
	l.calledWithResource = resource
	return l.limit, l.result
}

type muxMockLimiter map[string]*mockLimiter

func (l muxMockLimiter) RateLimit(resource ratelimit.Resource) (bool, ratelimit.Result) {
	for _, class := range resource.Classes() {
		if limiter, ok := l[class]; ok {
			return limiter.RateLimit(resource)
		}
	}
	return false, ratelimit.Result{}
}
