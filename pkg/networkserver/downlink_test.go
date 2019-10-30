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

package networkserver_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/networkserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

// handleDownlinkTaskQueueTest runs a test suite on q.
func handleDownlinkTaskQueueTest(t *testing.T, q DownlinkTaskQueue) {
	a := assertions.New(t)

	ctx := test.Context()

	pbs := [...]ttnpb.EndDeviceIdentifiers{
		{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "test-app",
			},
			DeviceID: "test-dev",
		},
		{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "test-app2",
			},
			DeviceID: "test-dev",
		},
		{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "test-app2",
			},
			DeviceID: "test-dev2",
		},
	}

	type slot struct {
		ctx   context.Context
		id    ttnpb.EndDeviceIdentifiers
		t     time.Time
		errCh chan<- error
	}

	popCtx := context.WithValue(ctx, struct{}{}, nil)
	nextPop := make(chan struct{})
	slotCh := make(chan slot)
	errCh := make(chan error)
	go func() {
		for {
			<-nextPop
			select {
			case <-ctx.Done():
				return
			case errCh <- q.Pop(popCtx, func(ctx context.Context, id ttnpb.EndDeviceIdentifiers, t time.Time) error {
				errCh := make(chan error)
				slotCh <- slot{
					ctx:   ctx,
					id:    id,
					t:     t,
					errCh: errCh,
				}
				return <-errCh
			}):
			}
		}
	}()

	// Ensure the goroutine has started
	nextPop <- struct{}{}

	select {
	case s := <-slotCh:
		t.Fatalf("Pop called f on empty schedule, slot: %+v", s)

	case err := <-errCh:
		a.So(err, should.BeNil)
		t.Fatal("Pop returned on empty schedule")

	case <-time.After(test.Delay):
	}

	err := q.Add(ctx, pbs[0], time.Unix(0, 0), false)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	select {
	case s := <-slotCh:
		a.So(s.id, should.Resemble, pbs[0])
		a.So(s.t, should.Equal, time.Unix(0, 0))
		if !a.So(s.ctx, should.HaveParentContextOrEqual, popCtx) {
			t.Log(s.ctx == popCtx)
			t.Fatal(s.ctx)
		}
		s.errCh <- nil

	case err := <-errCh:
		a.So(err, should.BeNil)
		t.Fatal("Pop returned without calling f on non-empty schedule")

	case <-time.After(10 * Timeout):
		t.Fatal("Timed out waiting for Pop to call f")
	}

	select {
	case err := <-errCh:
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

	case <-time.After(Timeout):
		t.Fatal("Timed out waiting for Pop to return")
	}

	err = q.Add(ctx, pbs[0], time.Unix(0, 42), false)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	err = q.Add(ctx, pbs[1], time.Now().Add(time.Hour), false)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	err = q.Add(ctx, pbs[1], time.Now().Add(2*time.Hour), true)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	err = q.Add(ctx, pbs[1], time.Unix(13, 0), false)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	err = q.Add(ctx, pbs[1], time.Unix(42, 0), true)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	err = q.Add(ctx, pbs[2], time.Now().Add(42*time.Hour), true)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	err = q.Add(ctx, pbs[2], time.Unix(42, 42), true)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	expectSlot := func(t *testing.T, expectedID ttnpb.EndDeviceIdentifiers, expectedAt time.Time) {
		nextPop <- struct{}{}

		t.Helper()

		a := assertions.New(t)

		select {
		case s := <-slotCh:
			a.So(s.id, should.Resemble, expectedID)
			a.So(s.t, should.Equal, expectedAt)
			a.So(s.ctx, should.HaveParentContextOrEqual, popCtx)
			s.errCh <- nil

		case err := <-errCh:
			a.So(err, should.BeNil)
			t.Fatal("Pop returned without calling f on non-empty schedule")

		case <-time.After(Timeout):
			t.Fatal("Timed out waiting for Pop to call f")
		}

		select {
		case err := <-errCh:
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

		case <-time.After(Timeout):
			t.Fatal("Timed out waiting for Pop to return")
		}
	}

	expectSlot(t, pbs[0], time.Unix(0, 42))
	expectSlot(t, pbs[1], time.Unix(42, 0))
	expectSlot(t, pbs[2], time.Unix(42, 42))
}

func TestDownlinkTaskQueues(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name string
		New  func(t testing.TB) (q DownlinkTaskQueue, closeFn func() error)
		N    uint16
	}{
		{
			Name: "Redis",
			New:  NewRedisDownlinkTaskQueue,
			N:    8,
		},
	} {
		for i := 0; i < int(tc.N); i++ {
			t.Run(fmt.Sprintf("%s/%d", tc.Name, i), func(t *testing.T) {
				t.Parallel()
				q, closeFn := tc.New(t)
				if closeFn != nil {
					defer func() {
						if err := closeFn(); err != nil {
							t.Errorf("Failed to close downlink schedule: %s", err)
						}
					}()
				}
				t.Run("1st run", func(t *testing.T) { handleDownlinkTaskQueueTest(t, q) })
				if t.Failed() {
					t.Skip("Skipping 2nd run")
				}
				t.Run("2nd run", func(t *testing.T) { handleDownlinkTaskQueueTest(t, q) })
			})
		}
	}
}
