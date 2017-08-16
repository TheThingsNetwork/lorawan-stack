// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package config

import (
	"os"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

type configWithPath struct {
	ConfigPaths []string `name:"config" shorthand:"c" description:"The location of the config path"`
	DataDir     string   `name:"data-dir" description:"The location of the data dir"`
}

type configWithoutPath struct{}

func TestConfigPathHome(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_CONFIG_HOME", "")
	os.Setenv("TEST_CONFIG", "")

	defaults := &configWithPath{}

	config := InitializeWithDefaults("test", defaults)
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "[$HOME/.test.yml]")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.ConfigPaths, should.Resemble, []string{"/home/johndoe/.test.yml"})
}

func TestConfigPathXDGHome(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_CONFIG_HOME", "/home/johndoe/.config")
	os.Setenv("TEST_CONFIG", "")

	defaults := &configWithPath{}

	config := InitializeWithDefaults("test", defaults)
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "[$XDG_CONFIG_HOME/test/test.yml]")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.ConfigPaths, should.Resemble, []string{"/home/johndoe/.config/test/test.yml"})
}

func TestConfigPathDefault(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_CONFIG_HOME", "")
	os.Setenv("TEST_CONFIG", "")

	defaults := &configWithPath{
		ConfigPaths: []string{
			"/env/test.yml",
			"/quu/qux.yml",
		},
	}

	config := InitializeWithDefaults("test", defaults)
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "[/env/test.yml,/quu/qux.yml]")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.ConfigPaths, should.Resemble, []string{"/env/test.yml", "/quu/qux.yml"})
}

func TestConfigPathEnv(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_CONFIG_HOME", "/home/johndoe/.config")
	os.Setenv("TEST_CONFIG", "/foo/bar/baz.yml")

	defaults := &configWithPath{}

	config := InitializeWithDefaults("test", defaults)
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "[/foo/bar/baz.yml]")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.ConfigPaths, should.Resemble, []string{"/foo/bar/baz.yml"})
}

func TestConfigPathCliFlag(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_CONFIG_HOME", "/home/johndoe/.config")
	os.Setenv("TEST_CONFIG", "")

	defaults := &configWithPath{}

	config := InitializeWithDefaults("test", defaults)
	a.So(config, should.NotBeNil)

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "[$XDG_CONFIG_HOME/test/test.yml]")

	config.Parse("--config", "/foo/bar", "--config", "/quu/qux")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.ConfigPaths, should.Resemble, []string{"/foo/bar", "/quu/qux"})
}

func TestConfigPathDefine(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_CONFIG_HOME", "/home/johndoe/.config")
	os.Setenv("TEST_CONFIG", "")

	defaults := &configWithoutPath{}

	config := InitializeWithDefaults("test", defaults)
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "[$XDG_CONFIG_HOME/test/test.yml]")
}

func TestDataDirHome(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_DATA_HOME", "")
	os.Setenv("TEST_DATA_DIR", "")

	defaults := &configWithPath{}

	config := InitializeWithDefaults("test", defaults, WithDataDirFlag("data-dir"))
	a.So(config, should.NotBeNil)

	config.Parse()

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.DataDir, should.Resemble, "/home/johndoe/.test")
}

func TestDataDirXDGHome(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_DATA_HOME", "/home/johndoe/.data")
	os.Setenv("TEST_DATA_DIR", "")

	defaults := &configWithPath{}

	config := InitializeWithDefaults("test", defaults, WithDataDirFlag("data-dir"))
	a.So(config, should.NotBeNil)

	config.Parse()

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.DataDir, should.Resemble, "/home/johndoe/.data/test")
}

func TestDataDirDefault(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_DATA_HOME", "/home/johndoe/.data")
	os.Setenv("TEST_DATA_DIR", "")

	defaults := &configWithPath{
		DataDir: "/var/run/test",
	}

	config := InitializeWithDefaults("test", defaults, WithDataDirFlag("data-dir"))
	a.So(config, should.NotBeNil)

	config.Parse()

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.DataDir, should.Resemble, "/var/run/test")
}

func TestDataDirEnv(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_DATA_HOME", "/home/johndoe/.data")
	os.Setenv("TEST_DATA_DIR", "/foo/bar")

	defaults := &configWithPath{}

	config := InitializeWithDefaults("test", defaults, WithDataDirFlag("data-dir"))
	a.So(config, should.NotBeNil)

	config.Parse()

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.DataDir, should.Resemble, "/foo/bar")
}

func TestDataDirCliFlag(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_DATA_HOME", "/home/johndoe/.data")
	os.Setenv("TEST_DATA_DIR", "")

	defaults := &configWithPath{}

	config := InitializeWithDefaults("test", defaults, WithDataDirFlag("data-dir"))
	a.So(config, should.NotBeNil)

	config.Parse("--data-dir", "/mnt/data")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.DataDir, should.Resemble, "/mnt/data")
}

func TestDataDirDefine(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_CONFIG_HOME", "/home/johndoe/.config")
	os.Setenv("TEST_CONFIG", "")

	defaults := &configWithoutPath{}

	config := InitializeWithDefaults("test", defaults, WithDataDirFlag("data-dir"))
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("data-dir")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "$XDG_DATA_HOME/test")
}
