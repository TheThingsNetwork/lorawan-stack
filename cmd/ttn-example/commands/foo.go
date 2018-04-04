// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package commands

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

type fooConfig struct {
	component.Config `name:",squash"`
	Bar              string `name:"bar" description:"The bar flag"`
}

var (
	fooCommand = &cobra.Command{
		Use:   "foo",
		Short: "The foo subcommand",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("Running foo")

			cfg := new(fooConfig)
			err := mgr.Unmarshal(cfg)
			if err != nil {
				return err
			}

			return printYAML(cfg)
		},
	}
)

func init() {
	// add the command to the root command
	Root.AddCommand(fooCommand)

	// add foo-specific config definitions and defaults
	fooCommand.Flags().AddFlagSet(mgr.WithConfig(&fooConfig{
		Bar: "baz",
	}))
}

// printYAML prints the nested config struct.
func printYAML(in interface{}) error {
	bs, err := yaml.Marshal(in)
	if err != nil {
		return err
	}

	fmt.Print(string(bs))
	return nil
}
