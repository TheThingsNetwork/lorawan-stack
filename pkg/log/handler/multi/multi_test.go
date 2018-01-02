// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package multi

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/log/handler/memory"
	"github.com/smartystreets/assertions"
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
