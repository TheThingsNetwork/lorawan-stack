// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

import (
	"bufio"
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
)

func TestHandlerNewCLIColors(t *testing.T) {
	a := assertions.New(t)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	// COLORTERM= TERM= does not enable colors
	os.Setenv("COLORTERM", "")
	os.Setenv("TERM", "")

	a.So(NewCLI(w).UseColor, assertions.ShouldBeFalse)

	// COLORTERM=0 forces colors off
	os.Setenv("COLORTERM", "0")
	os.Setenv("TERM", "colorterm")

	a.So(NewCLI(w).UseColor, assertions.ShouldBeFalse)

	// TERM with correct substring turns colors on
	os.Setenv("COLORTERM", "")
	os.Setenv("TERM", "colorterm")

	a.So(NewCLI(w).UseColor, assertions.ShouldBeTrue)

	// TERM with correct substring turns colors on
	os.Setenv("COLORTERM", "")
	os.Setenv("TERM", "xterm")

	a.So(NewCLI(w).UseColor, assertions.ShouldBeTrue)

	// COLORTERM=1 turns colors on
	os.Setenv("COLORTERM", "1")
	os.Setenv("TERM", "")

	a.So(NewCLI(w).UseColor, assertions.ShouldBeTrue)

	// COLORTERM=1 turns colors on
	os.Setenv("COLORTERM", "1")
	os.Setenv("TERM", "")

	// but UseColor(false) turns it off again
	a.So(NewCLI(w, UseColor(false)).UseColor, assertions.ShouldBeFalse)

	// COLORTERM=1 turns colors off
	os.Setenv("COLORTERM", "0")
	os.Setenv("TERM", "")

	// but UseColor(true) turns it off again
	a.So(NewCLI(w, UseColor(true)).UseColor, assertions.ShouldBeTrue)
}

func TestHandlerHandleLog(t *testing.T) {
	a := assertions.New(t)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	handler := NewCLI(w, UseColor(false))

	err := handler.HandleLog(&entry{
		message: "Foo",
		level:   DebugLevel,
		time:    time.Now(),
		fields:  Fields("a", 10, "b", "bar", "c", false, "d", 33.4),
	})
	a.So(err, assertions.ShouldBeNil)

	str := " DEBUG Foo                                      a=10 b=bar c=false d=33.4\n"

	err = w.Flush()
	a.So(err, assertions.ShouldBeNil)
	a.So(b.String(), assertions.ShouldEqual, str)
}
