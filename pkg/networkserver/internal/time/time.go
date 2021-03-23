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

// Package time wraps "time" and allows for custom implementations of key functions.
package time

import "time"

type (
	Time     = time.Time
	Duration = time.Duration
)

const (
	Nanosecond  = time.Nanosecond
	Microsecond = time.Microsecond
	Millisecond = time.Millisecond
	Second      = time.Second
	Minute      = time.Minute
	Hour        = time.Hour
)

var (
	now   func() time.Time                       = time.Now
	after func(d time.Duration) <-chan time.Time = time.After
)

func SetNow(f func() time.Time) func() {
	old := now
	now = f
	return func() {
		now = old
	}
}

func SetAfter(f func(time.Duration) <-chan time.Time) func() {
	old := after
	after = f
	return func() {
		after = old
	}
}

func Now() time.Time {
	return now()
}

func After(d time.Duration) <-chan time.Time {
	return after(d)
}

func Until(t time.Time) time.Duration {
	return t.Sub(now())
}

func Unix(sec int64, nsec int64) time.Time {
	return time.Unix(sec, nsec)
}
