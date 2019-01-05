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

package multi

import (
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/log/handler/memory"
)

type Entry struct {
	message string
	level   log.Level
	time    time.Time
	fields  log.Fielder
}

func (e *Entry) Message() string {
	return e.message
}

func (e *Entry) Level() log.Level {
	return e.level
}

func (e *Entry) Timestamp() time.Time {
	return e.time
}

func (e *Entry) Fields() log.Fielder {
	return e.fields
}

func Test(t *testing.T) {
	a := assertions.New(t)

	A := memory.New()
	B := memory.New()

	err := New(A, B).HandleLog(&Entry{
		message: "foo",
		fields:  log.Fields(),
		time:    time.Now(),
		level:   log.DebugLevel,
	})

	a.So(err, assertions.ShouldBeNil)
	a.So(A.Entries, assertions.ShouldHaveLength, 1)
	a.So(B.Entries, assertions.ShouldHaveLength, 1)
}
