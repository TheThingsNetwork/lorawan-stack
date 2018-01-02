// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package config

import (
	"os"
	"path"
	"strings"
)

// WithConfigFileFlag is an option for the manager that automatically enables the config file flag
// and tries to infer it from the $HOME and $XDG_CONFIG_HOME environment variables.
// You can only use this option once, it will panic otherwise.
func WithConfigFileFlag(flag string) Option {
	return func(m *Manager) {
		if m.configFlag != "" {
			panic("The WithConfigFileFlag option should only be used once")
		}

		m.configFlag = flag

		// use the default from the config if set
		if def := m.viper.GetStringSlice(flag); def != nil {
			m.viper.SetDefault(flag, def)
			return
		}

		file := m.name + ".yml"
		dotfile := "." + file

		var envPaths []string
		var paths []string

		if pwd := os.Getenv("PWD"); pwd != "" {
			envPaths = []string{path.Join("$PWD", dotfile)}
			paths = []string{path.Join(pwd, dotfile)}
		}

		// check HOME
		if home := os.Getenv("HOME"); home != "" {
			envPaths = []string{path.Join("$HOME", dotfile)}
			paths = []string{path.Join(home, dotfile)}
		}

		// check XDG_CONFIG_HOME
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			envPaths = []string{path.Join("$XDG_CONFIG_HOME", m.name, file)}
			paths = []string{path.Join(xdg, m.name, file)}
		}

		m.defaultPaths = paths
		m.viper.SetDefault(flag, paths)

		// set the flag default
		f := m.flags.Lookup(flag)
		if f != nil {
			f.DefValue = "[" + strings.Join(envPaths, ",") + "]"
		} else {
			m.flags.StringSlice(flag, envPaths, "Location of the configuration file")
		}
	}
}

// WithDataDirFlag is an option for the manager that automatically enables the data directory config flag
// and tries to infer it from the $HOME and $XDG_DATA_HOME environment variables.
// You can only use this option once, it will panic otherwise.
func WithDataDirFlag(flag string) Option {
	return func(m *Manager) {
		if m.dataDirFlag != "" {
			panic("The WithDataDirFlag option should only be used once")
		}

		m.dataDirFlag = flag
		dataDir := "$PWD"

		// use the default from defaults
		def := m.viper.GetString(flag)

		// check $HOME
		if home := os.Getenv("HOME"); home != "" {
			m.viper.AddConfigPath(home)
			m.viper.AddConfigPath(path.Join(home, "."+m.name))
			m.viper.SetDefault(flag, path.Join(home, "."+m.name))
			dataDir = path.Join("$HOME", "."+m.name)
		}

		// check $XDG_DATA_HOME
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			m.viper.SetDefault(flag, path.Join(xdg, m.name))
			dataDir = path.Join("$XDG_DATA_HOME", m.name)
		}

		// use the default
		if def != "" {
			m.viper.SetDefault(flag, def)
			dataDir = def
		}

		// set the flag default
		f := m.flags.Lookup(flag)
		if f != nil {
			f.DefValue = dataDir
		} else {
			m.flags.String(flag, dataDir, "Location of data directory")
		}
	}
}
