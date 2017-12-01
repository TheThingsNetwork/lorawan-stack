// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package main

import (
	"fmt"
	"os"

	"github.com/TheThingsNetwork/ttn/cmd/shared"
	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

// Config is the type of configuration
type Config struct {
	config.Base `name:",squash"`
	Int         int    `name:"int" description:"An example int"`
	String      string `name:"string" description:"An example string"`
}

// SubConfig is the type of config for the sub command
type SubConfig struct {
	Config `name:",squash"`
	Bar    string `name:"bar" description:"The bar config flag"`
}

var (
	defaults = &Config{
		Base:   shared.DefaultBaseConfig,
		Int:    42,
		String: "foo",
	}
	mgr = config.InitializeWithDefaults("example", defaults)
	cmd = &cobra.Command{
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			err := mgr.ReadInConfig()
			if err != nil {
				fmt.Println("Could not read config file:", err)
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg := new(Config)
			err := mgr.Unmarshal(cfg)
			if err != nil {
				panic(err)
			}

			printYAML(cfg)
		},
	}
	sub = &cobra.Command{
		Use: "sub",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := new(SubConfig)
			err := mgr.Unmarshal(cfg)
			if err != nil {
				panic(err)
			}

			printYAML(cfg)
		},
	}
)

func init() {
	cmd.Flags().AddFlagSet(mgr.Flags())

	sub.Flags().AddFlagSet(mgr.WithConfig(&SubConfig{
		Bar: "baz",
	}))
	cmd.AddCommand(sub)
}

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
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
