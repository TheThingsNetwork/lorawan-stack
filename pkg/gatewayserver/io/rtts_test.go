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

package io

import (
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestRTTs(t *testing.T) {
	a := assertions.New(t)

	rtts := newRTTs(5)

	_, ok := rtts.Last()
	a.So(ok, should.BeFalse)

	rtts.Record(2 * time.Second)
	rtts.Record(4 * time.Second)
	last, ok := rtts.Last()
	a.So(ok, should.BeTrue)
	a.So(last, should.Equal, 4*time.Second)

	min, max, median, count := rtts.Stats()
	a.So(min, should.Equal, 2*time.Second)
	a.So(max, should.Equal, 4*time.Second)
	a.So(median, should.Equal, 3*time.Second)
	a.So(count, should.Equal, 2)

	rtts.Record(8 * time.Second)
	min, max, median, count = rtts.Stats()
	a.So(min, should.Equal, 2*time.Second)
	a.So(max, should.Equal, 8*time.Second)
	a.So(median, should.Equal, 4*time.Second)
	a.So(count, should.Equal, 3)

	rtts.Record(5 * time.Second)
	rtts.Record(5 * time.Second)
	rtts.Record(5 * time.Second)
	rtts.Record(5 * time.Second)
	rtts.Record(5 * time.Second)
	min, max, median, count = rtts.Stats()
	a.So(min, should.Equal, 5*time.Second)
	a.So(max, should.Equal, 5*time.Second)
	a.So(median, should.Equal, 5*time.Second)
	a.So(count, should.Equal, 5)
}
