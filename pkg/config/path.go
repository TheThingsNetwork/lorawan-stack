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
// and tries to infer it from the working directory, user home directory and user configuration directory.
// The Option can only be used once.
func WithConfigFileFlag(flag string) Option {
	return func(m *Manager) {
		if m.configFlag != "" {
			panic("WithConfigFileFlag option should only be used once")
		}

		m.configFlag = flag

		// Use the default from the config if set.
		if def := m.viper.GetStringSlice(flag); def != nil {
			m.viper.SetDefault(flag, def)
			return
		}

		file := m.name + ".yml"
		dotfile := "." + file

		var paths, envPaths []string
		if dir, err := os.Getwd(); err == nil {
			paths = append(paths, path.Join(dir, dotfile))
			envPaths = append(envPaths, dotfile)
		}
		if dir, err := os.UserHomeDir(); err == nil {
			paths = append(paths, path.Join(dir, dotfile))
			envPaths = append(envPaths, path.Join(dir, dotfile))
		}
		if dir, err := os.UserConfigDir(); err == nil {
			paths = append(paths, path.Join(dir, dotfile))
			envPaths = append(envPaths, path.Join(dir, dotfile))
		}

		m.defaultPaths = paths
		m.viper.SetDefault(flag, paths)

		if f := m.flags.Lookup(flag); f != nil {
			f.DefValue = "[" + strings.Join(envPaths, ",") + "]"
		} else {
			m.flags.StringSlice(flag, envPaths, "Location of the configuration file")
		}
	}
}
