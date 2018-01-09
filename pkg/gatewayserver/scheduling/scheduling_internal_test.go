// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package scheduling

import (
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestWindowDurationSum(t *testing.T) {
	a := assertions.New(t)

	startingTime := time.Now()

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
