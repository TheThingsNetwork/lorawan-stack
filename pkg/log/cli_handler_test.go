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
