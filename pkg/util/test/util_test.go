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

package test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestMust(t *testing.T) {
	for i, tc := range []struct {
		Value       interface{}
		Error       error
		ShouldPanic bool
	}{
		{
			42,
			nil,
			false,
		},
		{
			errors.New("42"),
			nil,
			false,
		},
		{
			(error)(nil),
			errors.New("test"),
			true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			fn := func() { Must(tc.Value, tc.Error) }

			if tc.ShouldPanic {
				a.So(fn, should.Panic)
			} else if a.So(fn, should.NotPanic) {
				v := Must(tc.Value, tc.Error)
				a.So(v, should.Resemble, tc.Value)
			}
		})
	}
}

func TestMustMultiple(t *testing.T) {
	for i, tc := range []struct {
		Values      []interface{}
		ShouldPanic bool
	}{
		{
			nil,
			true,
		},
		{
			[]interface{}{},
			true,
		},
		{
			[]interface{}{(error)(nil)},
			false,
		},
		{
			[]interface{}{errors.New("42")},
			true,
		},
		{
			[]interface{}{42, (error)(nil)},
			false,
		},
		{
			[]interface{}{errors.New("42"), nil},
			false,
		},
		{
			[]interface{}{(error)(nil), errors.New("test")},
			true,
		},
		{
			[]interface{}{(error)(nil), (error)(nil), errors.New("test")},
			true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			fn := func() { MustMultiple(tc.Values...) }

			if tc.ShouldPanic {
				a.So(fn, should.Panic)
			} else if a.So(fn, should.NotPanic) {
				vs := MustMultiple(tc.Values...)
				a.So(vs, should.Resemble, tc.Values[:len(tc.Values)-1])
			}
		})
	}
}

func TestWaitTimeout(t *testing.T) {
	for _, tc := range []struct {
		Timeout time.Duration
		OK      bool
	}{
		{
			Timeout: 10 * time.Millisecond,
			OK:      false,
		},
		{
			Timeout: 30 * time.Millisecond,
			OK:      true,
		},
	} {
		t.Run(fmt.Sprintf("%v", tc.Timeout), func(t *testing.T) {
			a := assertions.New(t)

			ok := WaitTimeout(tc.Timeout, func() { time.Sleep(20 * time.Millisecond) })
			a.So(ok, should.Equal, tc.OK)
		})
	}
}
