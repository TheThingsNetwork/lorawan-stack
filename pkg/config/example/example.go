// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package main

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/spf13/cobra"
)

// Config is the type of configuration
type Config struct {
	config.ServiceBase `name:",squash"`
	Int                int    `name:"int" description:"An example int"`
	String             string `name:"string" description:"An example string"`
}

var (
	mgr *config.Manager
	cfg = &Config{}
	cmd = &cobra.Command{
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			err := mgr.ReadInConfig()
			if err != nil {
				fmt.Println("Could not read config file:", err)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := mgr.Unmarshal(cfg)
			if err != nil {
				panic(err)
			}

			printConfig(cfg, "")
		},
	}
)

func init() {
	defaults := &Config{
		Int:    42,
		String: "foo",
	}

	mgr = config.Initialize("example", defaults)
	cmd.Flags().AddFlagSet(mgr.Flags())
}

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}

// printConfig prints the nested config struct.
func printConfig(in interface{}, prefix string) {
	v := reflect.ValueOf(in)

	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			printConfig(reflect.Indirect(v).Interface(), prefix)
		}
	case reflect.Struct:
		t := v.Type()
		n := t.NumField()
		for i := 0; i < n; i++ {
			val := v.Field(i)

			if v.Kind() == reflect.Ptr {
				val = reflect.Indirect(val)
			}

			switch val.Kind() {
			case reflect.Struct:
				fmt.Printf("%s%s\n", prefix, t.Field(i).Name)
				printConfig(v.Field(i).Interface(), prefix+"  ")
			default:
				m, err := json.Marshal(val.Interface())
				if err != nil {
					panic(err)
				}
				fmt.Printf("%s%s = %v\n", prefix, t.Field(i).Name, string(m))
			}
		}
	default:
		fmt.Printf("%s%v\n", prefix, in)
	}
}
