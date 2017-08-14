package config

import (
	"os"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

type EmbeddedConfig struct {
	EmbeddedString string `name:"embedded-string" description:"Some embedded string"`
}

type EmbeddedConfigPtr struct {
	EmbeddedString string `name:"embeddedptr-string" description:"Some embedded string"`
}

type NestedConfig struct {
	String string `name:"string" description:"a nested string"`
}

type example struct {
	EmbeddedConfig `name:",squash"`
	// *EmbeddedConfigPtr `name:",squash"`

	Bool      bool          `name:"bool" description:"A single bool"`
	Duration  time.Duration `name:"duration" description:"A single duration"`
	Float     float64       `name:"float" description:"A single float"`
	Int       int           `name:"int" description:"A single int"`
	String    string        `name:"string" description:"A single string"`
	Strings   []string      `name:"strings" description:"A couple of strings"`
	StringPtr *string       `name:"stringptr" description:"A string ptr"`

	Nested    NestedConfig  `name:"nested" description:"A nested struct"`
	NestedPtr *NestedConfig `name:"nestedptr" description:"A nested struct ptr"`
	NotUsed   string        `name:"-"`
}

var (
	str      = "foo"
	defaults = &example{
		Bool:      true,
		Duration:  2 * time.Second,
		Float:     33.56,
		Int:       42,
		String:    "foo",
		Strings:   []string{"quu", "qux"},
		StringPtr: &str,
		Nested: NestedConfig{
			String: "nested-foo",
		},
		NestedPtr: &NestedConfig{
			String: "nested-bar",
		},
	}
)

func TestNilConfig(t *testing.T) {
	a := assertions.New(t)
	config := Initialize("empty", nil)
	a.So(config, should.NotBeNil)
}

func TestInvalidConfig(t *testing.T) {
	a := assertions.New(t)
	config := Initialize("invalid", "invalid")
	a.So(config, should.NotBeNil)
}

func TestConfigDefaults(t *testing.T) {
	a := assertions.New(t)

	config := Initialize("test", defaults)
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

	config := Initialize("test", defaults)
	a.So(config, should.NotBeNil)

	settings := new(example)

	os.Setenv("TEST_BOOL", "false")
	os.Setenv("TEST_DURATION", "10m")
	os.Setenv("TEST_FLOAT", "-112.45")
	os.Setenv("TEST_INT", "345")
	os.Setenv("TEST_STRING", "bababa")
	os.Setenv("TEST_STRINGS", "x y z")
	os.Setenv("TEST_STRINGPTR", "yo")
	os.Setenv("TEST_NESTED_STRING", "mud")
	os.Setenv("TEST_NESTEDPTR_STRING", "mad")

	// parse no command line args
	config.Parse()

	// unmarshal
	err := config.Unmarshal(settings)
	a.So(err, should.BeNil)

	str := "yo"
	a.So(settings, should.Resemble, &example{
		Bool:      false,
		Duration:  10 * time.Minute,
		Float:     -112.45,
		Int:       345,
		String:    "bababa",
		Strings:   []string{"x", "y", "z"},
		StringPtr: &str,
		Nested: NestedConfig{
			String: "mud",
		},
		NestedPtr: &NestedConfig{
			String: "mad",
		},
	})
}

func TestConfigFlags(t *testing.T) {
	a := assertions.New(t)

	config := Initialize("test", defaults)
	a.So(config, should.NotBeNil)

	settings := new(example)

	os.Setenv("TEST_BOOL", "")
	os.Setenv("TEST_DURATION", "")
	os.Setenv("TEST_FLOAT", "")
	os.Setenv("TEST_INT", "")
	os.Setenv("TEST_STRING", "")
	os.Setenv("TEST_STRINGS", "")
	os.Setenv("TEST_STRINGPTR", "")
	os.Setenv("TEST_NESTED_STRING", "")
	os.Setenv("TEST_NESTEDPTR_STRING", "")

	// parse no command line args
	config.Parse(
		"--duration", "10m",
		"--float", "12.45",
		"--int", "345",
		"--string", "bababa",
		"--strings", "x",
		"--strings", "y",
		"--strings", "z",
		"--stringptr", "yo",
		"--nested.string", "mud",
		"--nestedptr.string", "mad",
	)

	// unmarshal
	err := config.Unmarshal(settings)
	a.So(err, should.BeNil)

	str := "yo"
	a.So(settings, should.Resemble, &example{
		Bool:      true,
		Duration:  10 * time.Minute,
		Float:     12.45,
		Int:       345,
		String:    "bababa",
		Strings:   []string{"x", "y", "z"},
		StringPtr: &str,
		Nested: NestedConfig{
			String: "mud",
		},
		NestedPtr: &NestedConfig{
			String: "mad",
		},
	})
}
