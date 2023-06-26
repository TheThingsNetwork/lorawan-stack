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

// example cobra command with config.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	yaml "gopkg.in/yaml.v2"
)

// Config contains the base configuration with shared flags.
type Config struct {
	config.Base `name:",squash"`
	Int         int    `name:"int" description:"An example int"`
	String      string `name:"string" description:"An example string"`
}

// SubConfig contains the flags for the sub command.
type SubConfig struct {
	Config `name:",squash"`
	Bar    string `name:"bar" description:"The bar config flag"`
}

var (
	// The mgr is a configuration manager for the example program, initialized
	// with default config values.
	mgr = config.InitializeWithDefaults("example", "example", &Config{
		Int:    42,
		String: "foo",
	})

	// The cfg will be the config shared by all commands.
	cfg = new(Config)

	root = &cobra.Command{
		// PersistentPreRunE runs always, so we can use it to read in config files
		// and parse the shared config.
		PersistentPreRunE: func(*cobra.Command, []string) error {
			err := mgr.ReadInConfig()
			if err != nil {
				return err
			}
			return mgr.Unmarshal(cfg)
		},
		RunE: func(*cobra.Command, []string) error {
			return printYAML(cfg)
		},
	}

	sub = &cobra.Command{
		Use: "sub",

		// RunE parses the sub config that is only relevant for this command.
		RunE: func(*cobra.Command, []string) error {
			cfg := new(SubConfig)
			err := mgr.Unmarshal(cfg)
			if err != nil {
				return err
			}
			return printYAML(cfg)
		},
	}
)

func init() {
	root.AddCommand(sub)
	root.Flags().AddFlagSet(mgr.Flags())
	sub.Flags().AddFlagSet(mgr.WithConfig(&SubConfig{
		Bar: "baz",
	}))
}

func main() {
	// execute the commands and print resulting errors.
	err := root.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// printYAML prints the nested config struct.
func printYAML(in any) error {
	bs, err := yaml.Marshal(in)
	if err != nil {
		return err
	}

	fmt.Print(string(bs))
	return nil
}
