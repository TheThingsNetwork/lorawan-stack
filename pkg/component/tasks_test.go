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

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func init() {
	defaultTaskBackoff = [...]time.Duration{
		1 * test.Delay,
		2 * test.Delay,
		3 * test.Delay,
		4 * test.Delay,
	}
	backoffResetTime = 2 * test.Delay
}

func TestTasks(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	c, err := New(test.GetLogger(t), &Config{})
	a.So(err, should.BeNil)

	// Register a one-off task.
	oneOffWg := sync.WaitGroup{}
	oneOffWg.Add(1)
	c.RegisterTask(ctx, "one_off", func(_ context.Context) error {
		oneOffWg.Done()
		return nil
	}, TaskRestartNever)
	// Register an always restarting task.
	restartingWg := sync.WaitGroup{}
	restartingWg.Add(5)
	i := 0
	c.RegisterTask(ctx, "restarts", func(_ context.Context) error {
		i++
		if i <= 5 {
			restartingWg.Done()
		}
		return nil
	}, TaskRestartAlways)

	// Register a task that restarts on failure.
	failingWg := sync.WaitGroup{}
	failingWg.Add(1)
	j := 0
	c.RegisterTask(ctx, "restarts_on_failure", func(_ context.Context) error {
		j++
		if j < 5 {
			return errors.New("failed")
		}
		failingWg.Done()
		return nil
	}, TaskRestartOnFailure)

	// Wait for all invocations.
	test.Must(nil, c.Start())
	defer c.Close()
	oneOffWg.Wait()
	restartingWg.Wait()
	failingWg.Wait()
}

func TestTaskBackoffReset(t *testing.T) {
	if enabled := os.Getenv("TEST_TIMING"); enabled == "" {
		t.Skip("Timing sensitive tests are disabled")
	}

	a := assertions.New(t)
	ctx := test.Context()

	c, err := New(test.GetLogger(t), &Config{})
	a.So(err, should.BeNil)

	var wg sync.WaitGroup
	wg.Add(4)
	calls := 0
	lastCallTime := time.Time{}
	c.RegisterTask(ctx, "failing_but_recovering", func(_ context.Context) error {
		defer wg.Done()
		calls++
		backoff := time.Now().Sub(lastCallTime)
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
	}, TaskRestartAlways)

	// Wait for all invocations.
	test.Must(nil, c.Start())
	defer c.Close()
	wg.Wait()
}

func timeMult(t time.Duration, c float64) time.Duration {
	ft := float64(t)
	ft = ft * c
	return time.Duration(ft)
}
