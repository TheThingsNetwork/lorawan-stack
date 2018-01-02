// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/smartystreets/assertions"
)

func TestTestingHandler(t *testing.T) {
	a := assertions.New(t)

	handler := NewTestingHandler(t)

	err := handler.HandleLog(&Entry{
		M: "Foo",
		L: log.DebugLevel,
		T: time.Now(),
		F: log.Fields("a", 10, "b", "bar", "c", false, "d", 33.4),
	})
	a.So(err, assertions.ShouldBeNil)
}
