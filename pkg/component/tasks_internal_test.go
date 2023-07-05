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

package component

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var TestTaskBackoffConfig = &task.BackoffConfig{
	Jitter: task.DefaultBackoffJitter,
	IntervalFunc: task.MakeBackoffIntervalFunc(false, 2*test.Delay,
		test.Delay,
		2*test.Delay,
		3*test.Delay,
		4*test.Delay,
	),
}

func TestTaskBackoffReset(t *testing.T) {
	if enabled := os.Getenv("TEST_TIMING"); enabled == "" {
		t.Skip("Timing sensitive tests are disabled")
	}

	a := assertions.New(t)
	ctx := test.Context()

	c, err := New(test.GetLogger(t), &Config{})
	a.So(err, should.BeNil)

	var (
		calls        uint
		lastCallTime time.Time
		wg           sync.WaitGroup
	)
	wg.Add(4)
	c.RegisterTask(&task.Config{
		Context: ctx,
		ID:      "failing_but_recovering",
		Func: func(_ context.Context) error {
			defer wg.Done()
			calls++
			backoff := time.Since(lastCallTime)
			switch calls {
			case 1:
				// Nothing to expect in the initial call
			case 2:
				// Expect the first backoff interval
				a.So(backoff, should.BeBetweenOrEqual, timeMult(test.Delay, 0.9), timeMult(test.Delay, 1.1))
			case 3:
				// Same jitter requirement, but now act as if we are 'working'
				a.So(backoff, should.BeBetweenOrEqual, timeMult(2*test.Delay, 0.9), timeMult(2*test.Delay, 1.1))
				time.Sleep(3 * test.Delay)
			case 4:
				// Expect the backoff to have been reset
				a.So(backoff, should.BeBetweenOrEqual, timeMult(test.Delay, 0.9), timeMult(test.Delay, 1.1))
			}
			lastCallTime = time.Now()
			return nil
		},
		Restart: task.RestartAlways,
		Backoff: TestTaskBackoffConfig,
	})

	// Wait for all invocations.
	test.Must[any](nil, c.Start())
	defer c.Close()
	wg.Wait()
}

func timeMult(t time.Duration, c float64) time.Duration {
	ft := float64(t)
	ft = ft * c
	return time.Duration(ft)
}
