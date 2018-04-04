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

package scheduling

import (
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestWindowDurationSum(t *testing.T) {
	a := assertions.New(t)

	startingTime := SystemTime(time.Now())

	spans := []Span{
		{
			Start:    startingTime.Add(-1 * time.Second),
			Duration: 2 * time.Second,
		},
		{
			Start:    startingTime.Add(2 * time.Second),
			Duration: 2 * time.Second,
		},
		{
			Start:    startingTime.Add(-1 * time.Second),
			Duration: time.Second,
		},
		{
			Start:    startingTime.Add(time.Minute),
			Duration: time.Second,
		},
	}
	durationSum := sumWithinInterval(spans, startingTime, startingTime.Add(3*time.Second))
	a.So(durationSum, should.Equal, 2*time.Second)
}
