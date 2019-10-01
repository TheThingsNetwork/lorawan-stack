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

		// check PWD
		if pwd := os.Getenv("PWD"); pwd != "" {
			envPaths = append(envPaths, path.Join("$PWD", dotfile))
			paths = append(paths, path.Join(pwd, dotfile))
		}

		// check HOME
		if home := os.Getenv("HOME"); home != "" {
			envPaths = append(envPaths, path.Join("$HOME", dotfile))
			paths = append(paths, path.Join(home, dotfile))
		}

		// check XDG_CONFIG_HOME
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			envPaths = append(envPaths, path.Join("$XDG_CONFIG_HOME", m.name, file))
			paths = append(paths, path.Join(xdg, m.name, file))
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
