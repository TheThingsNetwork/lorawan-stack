// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
