package config

import (
	"os"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

type configWithPath struct {
	ConfigPath string `name:"config" shorthand:"c" description:"The location of the config path"`
	DataDir    string `name:"data-dir" description:"The location of the data dir"`
}

type configWithoutPath struct{}

func TestConfigPathHome(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_CONFIG_HOME", "")
	os.Setenv("TEST_CONFIG", "")

	defaults := &configWithPath{}

	config := Initialize("test", defaults)
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "$HOME/.test.yml")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.ConfigPath, should.Resemble, "/home/johndoe/.test.yml")
}

func TestConfigPathXDGHome(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_CONFIG_HOME", "/home/johndoe/.config")
	os.Setenv("TEST_CONFIG", "")

	defaults := &configWithPath{}

	config := Initialize("test", defaults)
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "$XDG_CONFIG_HOME/test/test.yml")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.ConfigPath, should.Resemble, "/home/johndoe/.config/test/test.yml")
}

func TestConfigPathDefault(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_CONFIG_HOME", "")
	os.Setenv("TEST_CONFIG", "")

	defaults := &configWithPath{
		ConfigPath: "/env/test.yml",
	}

	config := Initialize("test", defaults)
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "/env/test.yml")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.ConfigPath, should.Resemble, "/env/test.yml")
}

func TestConfigPathEnv(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_CONFIG_HOME", "/home/johndoe/.config")
	os.Setenv("TEST_CONFIG", "/foo/bar/baz.yml")

	defaults := &configWithPath{}

	config := Initialize("test", defaults)
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "/foo/bar/baz.yml")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.ConfigPath, should.Resemble, "/foo/bar/baz.yml")
}

func TestConfigPathCliFlag(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_CONFIG_HOME", "/home/johndoe/.config")
	os.Setenv("TEST_CONFIG", "")

	defaults := &configWithPath{
		ConfigPath: "",
	}

	config := Initialize("test", defaults)
	a.So(config, should.NotBeNil)

	config.Parse("--config", "/foo/bar")

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "$XDG_CONFIG_HOME/test/test.yml")

	res := new(configWithPath)
	config.Unmarshal(res)

	a.So(res.ConfigPath, should.Resemble, "/foo/bar")
}

func TestConfigPathDefine(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_CONFIG_HOME", "/home/johndoe/.config")
	os.Setenv("TEST_CONFIG", "")

	defaults := &configWithoutPath{}

	config := Initialize("test", defaults)
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("config")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "$XDG_CONFIG_HOME/test/test.yml")
}

func TestDataDirHome(t *testing.T) {
	a := assertions.New(t)

	os.Setenv("HOME", "/home/johndoe")
	os.Setenv("XDG_DATA_HOME", "")
	os.Setenv("TEST_DATA_DIR", "")

	defaults := &configWithPath{}

	config := Initialize("test", defaults, DataDir("data-dir"))
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

	config := Initialize("test", defaults, DataDir("data-dir"))
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

	config := Initialize("test", defaults, DataDir("data-dir"))
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

	config := Initialize("test", defaults, DataDir("data-dir"))
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

	config := Initialize("test", defaults, DataDir("data-dir"))
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

	config := Initialize("test", defaults, DataDir("data-dir"))
	a.So(config, should.NotBeNil)

	config.Parse()

	f := config.Flags().Lookup("data-dir")
	a.So(f, should.NotBeNil)
	a.So(f.DefValue, should.Resemble, "$XDG_DATA_HOME/test")
}
