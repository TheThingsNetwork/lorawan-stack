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

package redis_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var Timeout = 10 * test.Delay

func TestAddTask(t *testing.T) {
	a := assertions.New(t)

	cl, flush := test.NewRedis(t, "redis_test")
	defer flush()
	defer cl.Close()

	err := AddTask(cl, cl.Key("testKey"), 10, "testPayload", time.Unix(0, 42))
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	rets, err := cl.Client.XRead(&redis.XReadArgs{
		Streams: []string{InputTaskKey(cl.Key("testKey")), "0"},
		Count:   10,
		Block:   -1,
	}).Result()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	if a.So(rets, should.HaveLength, 1) {
		a.So(rets[0].Stream, should.Equal, InputTaskKey(cl.Key("testKey")))
		if a.So(rets[0].Messages, should.HaveLength, 1) {
			msg := rets[0].Messages[0]
			a.So(msg, should.Resemble, redis.XMessage{
				ID: msg.ID,
				Values: map[string]interface{}{
					"start_at": fmt.Sprintf("%d", time.Unix(0, 42).UnixNano()),
					"payload":  "testPayload",
				},
			})
		}
	}
}

func TestDispatchTasks(t *testing.T) {
	a := assertions.New(t)

	cl, flush := test.NewRedis(t, "redis_test")
	defer flush()
	defer cl.Close()

	now := time.Now()
	nextMin := now.Add(time.Hour)

	for _, k := range []string{
		"testKey",
		"testKey2",
	} {
		_, err := cl.XGroupCreateMkStream(InputTaskKey(cl.Key(k)), cl.Key("testGroup"), "0").Result()
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		_, err = cl.XGroupCreateMkStream(ReadyTaskKey(cl.Key(k)), cl.Key("testGroup"), "0").Result()
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
	}

	for _, x := range []*redis.XAddArgs{
		{
			Stream: InputTaskKey(cl.Key("testKey")),
			Values: map[string]interface{}{
				"start_at": fmt.Sprintf("%d", time.Unix(0, 42).UnixNano()),
				"payload":  "testPayload",
			},
		},
		{
			Stream: InputTaskKey(cl.Key("testKey")),
			Values: map[string]interface{}{
				"start_at": fmt.Sprintf("%d", time.Unix(0, 43).UnixNano()),
				"payload":  "testPayload",
			},
		},
		{
			Stream: InputTaskKey(cl.Key("testKey")),
			Values: map[string]interface{}{
				"start_at": fmt.Sprintf("%d", time.Unix(0, 41).UnixNano()),
				"payload":  "testPayload",
			},
		},
		{
			Stream: InputTaskKey(cl.Key("testKey")),
			Values: map[string]interface{}{
				"start_at": fmt.Sprintf("%d", time.Unix(0, 41).UnixNano()),
				"payload":  "testPayload",
			},
		},
		{
			Stream: InputTaskKey(cl.Key("testKey2")),
			Values: map[string]interface{}{
				"start_at": fmt.Sprintf("%d", time.Unix(0, 66).UnixNano()),
				"payload":  "testPayload",
			},
		},
		{
			Stream: InputTaskKey(cl.Key("testKey")),
			Values: map[string]interface{}{
				"start_at": nextMin.UnixNano(),
				"payload":  "testPayload2",
			},
		},
		{
			Stream: InputTaskKey(cl.Key("testKey")),
			Values: map[string]interface{}{
				"start_at": "0",
				"payload":  "testPayload3",
			},
		},
		{
			Stream: InputTaskKey(cl.Key("testKeyUnrelated")),
			Values: map[string]interface{}{
				"start_at": "0",
				"payload":  "testPayloadUnrelated",
			},
		},
	} {
		_, err := cl.Client.XAdd(x).Result()
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
	}

	min, err := DispatchTasks(cl, cl.Key("testGroup"), "testID", 100, time.Now(), cl.Key("testKey"), cl.Key("testKey2"))
	if !a.So(err, should.BeNil) {
		t.Errorf("Error cause: %v", errors.Cause(err))
		t.FailNow()
	}
	// NOTE: the timestamp gets converted to float64 under the hood, so some precision can be lost.
	a.So(min.UTC(), should.Resemble, time.Unix(0, int64(float64(nextMin.UnixNano()))).UTC())

	rets, err := cl.XReadStreams(ReadyTaskKey(cl.Key("testKey")), ReadyTaskKey(cl.Key("testKey2")), "0", "0").Result()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	if !a.So(rets, should.HaveLength, 2) || !a.So(rets[0].Messages, should.HaveLength, 2) || !a.So(rets[1].Messages, should.HaveLength, 1) {
		t.FailNow()
	}

	expected := []redis.XStream{
		{
			Stream: ReadyTaskKey(cl.Key("testKey")),
			Messages: []redis.XMessage{
				{
					Values: map[string]interface{}{
						"start_at": "0",
						"payload":  "testPayload3",
					},
				},
				{
					Values: map[string]interface{}{
						"start_at": fmt.Sprintf("%d", time.Unix(0, 42).UnixNano()),
						"payload":  "testPayload",
					},
				},
			},
		},
		{
			Stream: ReadyTaskKey(cl.Key("testKey2")),
			Messages: []redis.XMessage{
				{
					Values: map[string]interface{}{
						"start_at": fmt.Sprintf("%d", time.Unix(0, 66).UnixNano()),
						"payload":  "testPayload",
					},
				},
			},
		},
	}

	for i, ret := range rets {
		for j, msg := range ret.Messages {
			expected[i].Messages[j].ID = msg.ID
		}
	}
	a.So(rets, should.Resemble, expected)
}

func TestPopTask(t *testing.T) {
	a := assertions.New(t)

	cl, flush := test.NewRedis(t, "redis_test")
	defer flush()
	defer cl.Close()

	ctx, cancel := context.WithCancel(test.Context())
	defer cancel()

	err := PopTask(cl, cl.Key("testGroup"), "testID", -1, func(string, string, time.Time) error {
		t.Fatal("f must not be called")
		return nil
	}, cl.Key("testKeyNonExistent"))
	a.So(err, should.NotBeNil)

	for _, k := range []string{
		"testKey",
		"testKey2",
		"testKey3", // Stays empty
	} {
		_, err := cl.XGroupCreateMkStream(ReadyTaskKey(cl.Key(k)), cl.Key("testGroup"), "0").Result()
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
	}

	err = PopTask(cl, "testGroupNonExistent", "testID", -1, func(string, string, time.Time) error {
		t.Fatal("f must not be called")
		return nil
	}, cl.Key("testKey"))
	a.So(err, should.NotBeNil)

	err = PopTask(cl, cl.Key("testGroup"), "testID", -1, func(string, string, time.Time) error {
		t.Fatal("f must not be called")
		return nil
	})
	a.So(err, should.BeNil)

	err = PopTask(cl, cl.Key("testGroup"), "testID", -1, func(string, string, time.Time) error {
		t.Fatal("f must not be called")
		return nil
	}, cl.Key("testKey"), cl.Key("testKey2"))
	a.So(err, should.BeNil)

	err = PopTask(cl, cl.Key("testGroup"), "testID", -1, func(string, string, time.Time) error {
		t.Fatal("f must not be called")
		return nil
	}, cl.Key("testKey2"))
	a.So(err, should.BeNil)

	err = PopTask(cl, cl.Key("testGroup"), "testID", -1, func(string, string, time.Time) error {
		t.Fatal("f must not be called")
		return nil
	}, cl.Key("testKeyNonExistent"))
	a.So(err, should.NotBeNil)

	go func() {
		err := PopTask(cl, cl.Key("testGroup"), "testID", 0, func(string, string, time.Time) error {
			t.Fatal("f must not be called")
			return nil
		}, cl.Key("testKey3"))
		select {
		case <-ctx.Done():
			t.Logf("Blocked PopTask returned, error: %s", err)
		default:
			t.Fatalf("PopTask returned too early, error: %s", err)
		}
	}()

	for _, x := range []*redis.XAddArgs{
		{
			Stream: ReadyTaskKey(cl.Key("testKey")),
			Values: map[string]interface{}{
				"start_at": fmt.Sprintf("%d", time.Unix(0, 42).UnixNano()),
				"payload":  "testPayload",
			},
		},
		{
			Stream: ReadyTaskKey(cl.Key("testKey")),
			Values: map[string]interface{}{
				"start_at": fmt.Sprintf("%d", time.Unix(0, 43).UnixNano()),
				"payload":  "testPayload2",
			},
		},
		{
			Stream: ReadyTaskKey(cl.Key("testKey")),
			Values: map[string]interface{}{
				"start_at": fmt.Sprintf("%d", time.Unix(0, 41).UnixNano()),
				"payload":  "testPayload",
			},
		},
		{
			Stream: ReadyTaskKey(cl.Key("testKey2")),
			Values: map[string]interface{}{
				"start_at": fmt.Sprintf("%d", time.Unix(0, 41).UnixNano()),
				"payload":  "testPayload",
			},
		},
	} {
		_, err := cl.Client.XAdd(x).Result()
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
	}

	fCalls := 0
	err = PopTask(cl, cl.Key("testGroup"), "testID", -1, func(k string, payload string, startAt time.Time) error {
		a.So(fCalls, should.Equal, 0)
		a.So(k, should.Equal, cl.Key("testKey"))
		a.So(payload, should.Equal, "testPayload")
		a.So(startAt, should.Equal, time.Unix(0, 42).UTC())
		fCalls++
		return nil
	}, cl.Key("testKey"))
	a.So(err, should.BeNil)

	err = PopTask(cl, cl.Key("testGroup"), "testID2", -1, func(k string, payload string, startAt time.Time) error {
		a.So(fCalls, should.Equal, 1)
		a.So(k, should.Equal, cl.Key("testKey"))
		a.So(payload, should.Equal, "testPayload2")
		a.So(startAt, should.Equal, time.Unix(0, 43).UTC())
		fCalls++
		return nil
	}, cl.Key("testKey"))
	a.So(err, should.BeNil)

	errTest := errors.New("test")
	err = PopTask(cl, cl.Key("testGroup"), "testID", -1, func(k string, payload string, startAt time.Time) error {
		a.So(fCalls, should.Equal, 2)
		a.So(k, should.Equal, cl.Key("testKey"))
		a.So(payload, should.Equal, "testPayload")
		a.So(startAt, should.Equal, time.Unix(0, 41).UTC())
		fCalls++
		return errTest
	}, cl.Key("testKey"))
	a.So(err, should.Resemble, errTest)

	err = PopTask(cl, cl.Key("testGroup"), "testID", -1, func(k string, payload string, startAt time.Time) error {
		a.So(fCalls, should.Equal, 3)
		a.So(k, should.Equal, cl.Key("testKey2"))
		a.So(payload, should.Equal, "testPayload")
		a.So(startAt, should.Equal, time.Unix(0, 41).UTC())
		fCalls++
		return nil
	}, cl.Key("testKey2"))
	a.So(err, should.BeNil)

	a.So(fCalls, should.Equal, 4)

	xp, err := cl.XPending(ReadyTaskKey(cl.Key("testKey")), cl.Key("testGroup")).Result()
	if !a.So(err, should.BeNil) {
		t.Fatal("Failed to query pending tasks")
	}
	a.So(xp.Lower, should.Equal, xp.Higher)
	a.So(xp, should.Resemble, &redis.XPending{
		Count:  1,
		Lower:  xp.Lower,
		Higher: xp.Higher,
		Consumers: map[string]int64{
			"testID": 1,
		},
	})

	xp, err = cl.XPending(ReadyTaskKey(cl.Key("testKey2")), cl.Key("testGroup")).Result()
	if !a.So(err, should.BeNil) {
		t.Fatal("Failed to query pending tasks")
	}
	a.So(xp, should.Resemble, &redis.XPending{})
}

func TestInitTaskGroup(t *testing.T) {
	for _, tc := range []struct {
		Name           string
		Populate       func(t *testing.T, cl *Client) bool
		Group, Key     string
		ErrorAssertion func(t *testing.T, err error) bool
	}{
		{
			Name:           "no streams/no groups",
			Populate:       func(t *testing.T, cl *Client) bool { return true },
			Group:          "testGroup",
			Key:            "testKey",
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "streams exist/groups exist",
			Populate: func(t *testing.T, cl *Client) bool {
				a := assertions.New(t)
				_, err := cl.XGroupCreateMkStream(InputTaskKey(cl.Key("testKey")), cl.Key("testGroup"), "0").Result()
				if !a.So(err, should.BeNil) {
					return false
				}
				_, err = cl.XGroupCreateMkStream(ReadyTaskKey(cl.Key("testKey")), cl.Key("testGroup"), "0").Result()
				if !a.So(err, should.BeNil) {
					return false
				}
				return true
			},
			Group:          "testGroup",
			Key:            "testKey",
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a := assertions.New(t)

			cl, flush := test.NewRedis(t, "redis_test")
			defer flush()
			defer cl.Close()

			a.So(tc.Populate(t, cl), should.BeTrue)

			err := InitTaskGroup(cl, cl.Key(tc.Group), cl.Key(tc.Key))
			a.So(tc.ErrorAssertion(t, err), should.BeTrue)
		})
	}

}

func TestTaskQueue(t *testing.T) {
	a := assertions.New(t)

	cl, flush := test.NewRedis(t, "redis_test")
	defer flush()
	defer cl.Close()

	q := &TaskQueue{
		Redis:  cl,
		MaxLen: 42,
		Group:  "testGroup",
		ID:     "testID",
		Key:    cl.Key("test"),
	}

	err := q.Init()
	a.So(err, should.BeNil)

	runCtx, cancel := context.WithCancel(test.Context())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	runErrCh := make(chan error, 1)
	go func() {
		wg.Done()
		t.Log("Running Redis downlink task queue...")
		err := q.Run(runCtx)
		runErrCh <- err
		close(runErrCh)
		t.Logf("Stopped Redis downlink task queue with error: %s", err)
	}()
	wg.Wait()

	type task struct {
		payload string
		startAt time.Time
		errCh   chan<- error
	}
	type popReq struct {
		ctx    context.Context
		taskCh chan task
		errCh  chan error
	}
	newPopReq := func(ctx context.Context) popReq {
		return popReq{
			ctx:    ctx,
			taskCh: make(chan task, 1),
			errCh:  make(chan error, 1),
		}
	}
	popReqCh := make(chan popReq)
	go func() {
		for req := range popReqCh {
			req.errCh <- q.Pop(req.ctx, func(payload string, startAt time.Time) error {
				errCh := make(chan error, 1)
				req.taskCh <- task{
					payload: payload,
					startAt: startAt,
					errCh:   errCh,
				}
				close(req.taskCh)
				return <-errCh
			})
			close(req.errCh)
		}
	}()

	ctx, _ := context.WithDeadline(test.Context(), time.Now().Add(-1))
	req := newPopReq(ctx)
	popReqCh <- req

	select {
	case x := <-req.taskCh:
		t.Fatalf("Non-blocking Pop called f on empty schedule, task: %+v", x)

	case err := <-req.errCh:
		a.So(err, should.BeNil)

	case <-time.After(Timeout):
		t.Fatalf("Non-blocking Pop blocked for %v", Timeout)
	}

	req = newPopReq(test.Context())
	popReqCh <- req

	select {
	case x := <-req.taskCh:
		t.Fatalf("Blocking Pop called f on empty schedule, task: %+v", x)

	case err := <-req.errCh:
		a.So(err, should.BeNil)
		t.Fatal("Blocking Pop returned on empty schedule")

	case <-time.After(10 * test.Delay):
	}

	err = q.Add("testPayload", time.Unix(0, 0))
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	select {
	case x := <-req.taskCh:
		a.So(x.payload, should.Equal, "testPayload")
		a.So(x.startAt, should.Equal, time.Unix(0, 0))
		x.errCh <- nil
		close(x.errCh)

	case err := <-req.errCh:
		a.So(err, should.BeNil)
		t.Fatal("Run returned without calling f on non-empty schedule")

	case <-time.After(Timeout):
		t.Fatal("Timed out waiting for Pop to call f")
	}

	err = q.Add("0", time.Unix(0, 42))
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	err = q.Add("1", time.Unix(42, 0))
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	err = q.Add("2", time.Unix(42, 42))
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	expectTask := func(t *testing.T, expectedPayload string, expectedAt time.Time) {
		t.Helper()

		a := assertions.New(t)

		req := newPopReq(test.Context())
		popReqCh <- req

		select {
		case x := <-req.taskCh:
			a.So(x.payload, should.Equal, expectedPayload)
			a.So(x.startAt, should.Equal, expectedAt)
			x.errCh <- nil
			close(x.errCh)

		case err := <-req.errCh:
			a.So(err, should.BeNil)
			t.Fatal("Pop returned without calling f on non-empty schedule")

		case <-time.After(Timeout):
			t.Fatal("Timed out waiting for Pop to call f")
		}

		select {
		case err := <-req.errCh:
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

		case <-time.After(Timeout):
			t.Fatal("Timed out waiting for Pop to return")
		}
	}

	expectTask(t, "0", time.Unix(0, 42))
	expectTask(t, "1", time.Unix(42, 0))
	expectTask(t, "2", time.Unix(42, 42))

	cancel()
	// Unblock DispatchTasks in Run()
	err = q.Add("42", time.Unix(0, 0))
	a.So(err, should.BeNil)

	select {
	case err := <-runErrCh:
		if err != nil && err != context.Canceled {
			t.Errorf("Failed to run queue: %s(cause: %s)", err, errors.Cause(err))
		}

	case <-time.After(Timeout):
		t.Error("Timed out waiting for Run to return")
	}
}
