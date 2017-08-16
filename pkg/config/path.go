// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package config

import (
	"os"
	"path"
	"strings"
)

// WithConfigFileFlag is an option for the manager that automatically enables the config file flag
// and tries to infer it from the $HOME and $XDG_CONFIG_HOME environment variables.
func WithConfigFileFlag(flag string) Option {
	return func(m *Manager) {
		m.configFlag = flag

		configPaths := []string{path.Join("$PWD", "."+m.name+".yml")}
		def := m.viper.GetStringSlice(flag)

		// check HOME
		if home := os.Getenv("HOME"); home != "" {
			m.viper.AddConfigPath(home)
			m.viper.AddConfigPath(path.Join(home, "."+m.name))
			m.viper.SetDefault(flag, []string{path.Join(home, "."+m.name+".yml")})

			configPaths = []string{path.Join("$HOME", "."+m.name+".yml")}
		}

		// check XDG_CONFIG_HOME
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			m.viper.AddConfigPath(xdg)
			m.viper.SetDefault(flag, []string{path.Join(xdg, m.name, m.name+".yml")})
			configPaths = []string{path.Join("$XDG_CONFIG_HOME", m.name, m.name+".yml")}
		}

		// use the default
		if def != nil {
			configPaths = def
			for _, pth := range def {
				m.viper.AddConfigPath(pth)
			}

			m.viper.SetDefault(flag, def)
		}

		// set the flag default
		f := m.flags.Lookup(flag)
		if f != nil {
			f.DefValue = "[" + strings.Join(configPaths, ",") + "]"
		} else {
			m.flags.StringSlice(flag, configPaths, "Location of the configuration file")
		}
	}
}

// WithDataDirFlag is an option for the manager that automatically enables the data directory config flag
// and tries to infer it from the $HOME and $XDG_DATA_HOME environment variables.
func WithDataDirFlag(flag string) Option {
	return func(m *Manager) {
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
