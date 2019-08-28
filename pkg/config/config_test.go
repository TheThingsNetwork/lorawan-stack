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

package config

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

type EmbeddedConfig struct {
	EmbeddedString string `name:"embedded-string" description:"Some embedded string"`
}

type NestedConfig struct {
	String string `name:"string" description:"a nested string"`
}

type Custom int

func (c Custom) ConfigString() string {
	switch int(c) {
	case 42:
		return "foo"
	case 112:
		return "bar"
	default:
		return ""
	}
}

func (c *Custom) UnmarshalConfigString(text string) error {
	switch text {
	case "foo":
		*c = Custom(42)
		return nil
	case "bar":
		*c = Custom(112)
		return nil
	}
	return fmt.Errorf("Could not parse custom value %s", text)
}

type example struct {
	EmbeddedConfig `name:",squash"`

	Bool      bool          `name:"bool" description:"A single bool"`
	Duration  time.Duration `name:"duration" description:"A single duration"`
	Time      time.Time     `name:"time" description:"A single time"`
	TimePtr   *time.Time    `name:"timeptr" description:"A single time"`
	Float     float64       `name:"float" description:"A single float"`
	Int       int           `name:"int" description:"A single int"`
	String    string        `name:"string" shorthand:"s" description:"A single string"`
	Strings   []string      `name:"strings" description:"A couple of strings"`
	StringPtr *string       `name:"stringptr" description:"A string ptr"`
	Bytes     []byte        `name:"bytes" description:"A slice of bytes"`

	StringMap      map[string]string   `name:"stringmap" description:"A map of strings"`
	BufferMap      map[string][]byte   `name:"buffermap" description:"A map of buffers"`
	StringMapSlice map[string][]string `name:"stringmapslice" description:"A map of string slices"`

	Nested    NestedConfig  `name:"nested" description:"A nested struct"`
	NestedPtr *NestedConfig `name:"nestedptr" description:"A nested struct ptr"`
	NotUsed   string        `name:"-"`

	Custom  Custom   `name:"custom" description:"A custom type"`
	Customs []Custom `name:"customs" description:"A slice of custom types"`

	FileOnly interface{} `name:"file-only" file-only:"true"`
}

var (
	str      = "foo"
	defaults = &example{
		Bool:      true,
		Duration:  2 * time.Second,
		Time:      time.Date(1991, time.September, 12, 23, 24, 0, 0, time.UTC),
		Float:     33.56,
		Int:       42,
		String:    "foo",
		Strings:   []string{"quu", "qux"},
		StringPtr: &str,
		Bytes:     []byte{0x01, 0xFA},
		StringMap: map[string]string{
			"foo": "bar",
		},
		BufferMap: map[string][]byte{
			"foo": {0xb, 0xa, 0xf},
			"bar": {0xf, 0x0, 0x0},
		},
		StringMapSlice: map[string][]string{
			"foo": {"bar", "baz"},
			"quu": {"qux"},
		},
		Nested: NestedConfig{
			String: "nested-foo",
		},
		NestedPtr: &NestedConfig{
			String: "nested-bar",
		},
		Custom: Custom(42),
		Customs: []Custom{
			Custom(42),
			Custom(112),
		},
		FileOnly: 33,
	}
)

func TestNilConfig(t *testing.T) {
	a := assertions.New(t)
	config := InitializeWithDefaults("empty", "empty", nil)
	a.So(config, should.NotBeNil)
}

func TestInvalidConfig(t *testing.T) {
	a := assertions.New(t)
	config := InitializeWithDefaults("invalid", "invalid", struct{}{})
	a.So(config, should.NotBeNil)
}

func TestConfigDefaults(t *testing.T) {
	a := assertions.New(t)

	config := InitializeWithDefaults("test", "test", defaults)
	a.So(config, should.NotBeNil)

	settings := new(example)

	// parse no command line args
	config.Parse()

	// unmarshal
	err := config.Unmarshal(settings)
	a.So(err, should.BeNil)

	a.So(settings, should.Resemble, defaults)
}

func TestConfigEnv(t *testing.T) {
	a := assertions.New(t)

	config := InitializeWithDefaults("test", "test", defaults)
	a.So(config, should.NotBeNil)

	settings := new(example)

	os.Setenv("TEST_BOOL", "false")
	os.Setenv("TEST_DURATION", "10m")
	os.Setenv("TEST_TIME", "2017-08-12 01:02:03 +0000 UTC")
	os.Setenv("TEST_FLOAT", "-112.45")
	os.Setenv("TEST_INT", "345")
	os.Setenv("TEST_STRING", "bababa")
	os.Setenv("TEST_STRINGS", "x y z")
	os.Setenv("TEST_STRINGPTR", "yo")
	os.Setenv("TEST_BYTES", "FA00BB")
	os.Setenv("TEST_STRINGMAP", "q=r s=t")
	os.Setenv("TEST_BUFFERMAP", "a=0x0bcd c=0x0def")
	os.Setenv("TEST_STRINGMAPSLICE", "a=b a=c d=e")
	os.Setenv("TEST_NESTED_STRING", "mud")
	os.Setenv("TEST_NESTEDPTR_STRING", "mad")
	os.Setenv("TEST_CUSTOM", "bar")
	os.Setenv("TEST_CUSTOMS", "bar")

	// parse no command line args
	config.Parse()

	// unmarshal into struct
	err := config.Unmarshal(settings)
	a.So(err, should.BeNil)

	str := "yo"
	a.So(settings, should.Resemble, &example{
		Bool:      false,
		Duration:  10 * time.Minute,
		Time:      time.Date(2017, time.August, 12, 01, 02, 03, 0, time.UTC),
		Float:     -112.45,
		Int:       345,
		String:    "bababa",
		Strings:   []string{"x", "y", "z"},
		StringPtr: &str,
		Bytes:     []byte{0xFA, 0x00, 0xBB},
		StringMap: map[string]string{
			"q": "r",
			"s": "t",
		},
		BufferMap: map[string][]byte{
			"a": {0xb, 0xcd},
			"c": {0xd, 0xef},
		},
		StringMapSlice: map[string][]string{
			"a": {"b", "c"},
			"d": {"e"},
		},
		Nested: NestedConfig{
			String: "mud",
		},
		NestedPtr: &NestedConfig{
			String: "mad",
		},
		Custom: Custom(112),
		Customs: []Custom{
			Custom(112),
		},
		FileOnly: defaults.FileOnly,
	})
}

func TestBadConfigEnv(t *testing.T) {
	a := assertions.New(t)

	config := InitializeWithDefaults("test", "test", defaults)
	a.So(config, should.NotBeNil)

	settings := new(example)

	os.Setenv("TEST_STRINGMAP", "q=r ff = f ff s=t")
	os.Setenv("TEST_BUFFERMAP", "a=0x0b  cd c=0x0=def")
	os.Setenv("TEST_STRINGMAPSLICE", "a=b= a=c fzef f d=e")

	// parse no command line args
	config.Parse()

	// unmarshal into struct
	err := config.Unmarshal(settings)
	a.So(err, should.NotBeNil)
}

func TestConfigFlags(t *testing.T) {
	a := assertions.New(t)

	config := InitializeWithDefaults("test", "test", defaults)
	a.So(config, should.NotBeNil)

	settings := new(example)

	os.Unsetenv("TEST_BOOL")
	os.Unsetenv("TEST_DURATION")
	os.Unsetenv("TEST_TIME")
	os.Unsetenv("TEST_FLOAT")
	os.Unsetenv("TEST_INT")
	os.Unsetenv("TEST_STRING")
	os.Unsetenv("TEST_STRINGS")
	os.Unsetenv("TEST_STRINGPTR")
	os.Unsetenv("TEST_BYTES")
	os.Unsetenv("TEST_STRINGMAP")
	os.Unsetenv("TEST_BUFFERMAP")
	os.Unsetenv("TEST_STRINGMAPSLICE")
	os.Unsetenv("TEST_NESTED_STRING")
	os.Unsetenv("TEST_NESTEDPTR_STRING")
	os.Unsetenv("TEST_CUSTOM")

	// parse command line args
	config.Parse(
		"--duration", "10m",
		"--time", "2017-08-12 01:02:03 +0000 UTC",
		"--float", "12.45",
		"--int", "345",
		"--string", "bababa",
		"--strings", "x",
		"--strings", "y",
		"--strings", "z",
		"--stringptr", "yo",
		"--bytes", "99FD",
		"--stringmap", "q=r",
		"--stringmap", "s=t",
		"--buffermap", "a=0x0bcd",
		"--buffermap", "c=0x0def",
		"--nested.string", "mud",
		"--nestedptr.string", "mad",
		"--custom", "bar",
		"--customs", "bar",
		"--customs", "foo",
		"--stringmapslice", "a=b",
		"--stringmapslice", "a=c",
		"--stringmapslice", "d=e",
	)

	// unmarshal
	err := config.Unmarshal(settings)
	a.So(err, should.BeNil)

	str := "yo"
	a.So(settings, should.Resemble, &example{
		Bool:      true,
		Duration:  10 * time.Minute,
		Time:      time.Date(2017, time.August, 12, 01, 02, 03, 0, time.UTC),
		Float:     12.45,
		Int:       345,
		String:    "bababa",
		Strings:   []string{"x", "y", "z"},
		StringPtr: &str,
		Bytes:     []byte{0x99, 0xFD},
		StringMap: map[string]string{
			"q": "r",
			"s": "t",
		},
		BufferMap: map[string][]byte{
			"a": {0xb, 0xcd},
			"c": {0xd, 0xef},
		},
		StringMapSlice: map[string][]string{
			"a": {"b", "c"},
			"d": {"e"},
		},
		Nested: NestedConfig{
			String: "mud",
		},
		NestedPtr: &NestedConfig{
			String: "mad",
		},
		Custom: Custom(112),
		Customs: []Custom{
			Custom(112),
			Custom(42),
		},
		FileOnly: defaults.FileOnly,
	})
}

func TestBadConfigFlags(t *testing.T) {
	a := assertions.New(t)

	config := InitializeWithDefaults("test", "test", defaults)
	a.So(config, should.NotBeNil)
	settings := new(example)

	// parse command line args
	config.Parse(
		"--buffermap", "a=0x 0osvpz=bcd",
		"--stringmap", "q=r fhoez=fff",
	)

	// unmarshal
	err := config.Unmarshal(settings)
	a.So(err, should.NotBeNil)
}

func TestConfigShorthand(t *testing.T) {
	a := assertions.New(t)

	config := InitializeWithDefaults("test", "test", defaults)
	a.So(config, should.NotBeNil)

	settings := new(example)

	os.Setenv("TEST_STRING", "")

	// parse command line args
	config.Parse("-s", "bababa")

	// unmarshal
	err := config.Unmarshal(settings)
	a.So(err, should.BeNil)

	a.So(settings.String, should.Resemble, "bababa")
}
