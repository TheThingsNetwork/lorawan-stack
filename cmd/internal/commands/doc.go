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

package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

func disableAutoGenTag(cmd *cobra.Command) {
	cmd.DisableAutoGenTag = true
	for _, sub := range cmd.Commands() {
		disableAutoGenTag(sub)
	}
}

// GenManPages generates man pages for the given root command.
func GenManPages(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "gen-man-pages",
		Hidden: true,
		Short:  fmt.Sprintf("Generate man pages for %s", root.Name()),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("out")
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return err
				}
			}
			disableAutoGenTag(root)
			return doc.GenManTree(root, &doc.GenManHeader{
				Title:   strings.ToUpper(root.Name()),
				Section: "1",
				Manual:  root.Root().Short,
				Source:  "TTN",
			}, dir)
		},
	}
	cmd.Flags().StringP("out", "o", "doc", "output directory")
	return cmd
}

// GenMDDoc generates markdown documentation for the given root command.
func GenMDDoc(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "gen-md-doc",
		Hidden: true,
		Short:  fmt.Sprintf("Generate markdown documentation for %s", root.Name()),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("out")
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return err
				}
			}
			disableAutoGenTag(root)
			return doc.GenMarkdownTree(root, dir)
		},
	}
	cmd.Flags().StringP("out", "o", "doc", "output directory")
	return cmd
}

// GenYAMLDoc generates yaml documentation for the given root command.
func GenYAMLDoc(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "gen-yaml-doc",
		Hidden: true,
		Short:  fmt.Sprintf("Generate yaml documentation for %s", root.Name()),
		RunE: func(cmd *cobra.Command, args []string) error {
			type Flag struct {
				Name         string `yaml:"name,omitempty"`
				Type         string `yaml:"type,omitempty"`
				Shorthand    string `yaml:"shorthand,omitempty"`
				Usage        string `yaml:"usage,omitempty"`
				DefaultValue string `yaml:"default_value,omitempty"`
				Hidden       bool   `yaml:"hidden,omitempty"`
			}
			buildFlag := func(flag *pflag.Flag) *Flag {
				doc := &Flag{
					Name:         flag.Name,
					Type:         flag.Value.Type(),
					Shorthand:    flag.Shorthand,
					Usage:        flag.Usage,
					DefaultValue: flag.DefValue,
					Hidden:       flag.Hidden,
				}
				return doc
			}
			type Command struct {
				Path            string   `yaml:"path,omitempty"`
				ParentPath      string   `yaml:"parent_path,omitempty"`
				Name            string   `yaml:"name,omitempty"`
				Use             string   `yaml:"use,omitempty"`
				Aliases         []string `yaml:"aliases,omitempty"`
				Short           string   `yaml:"short,omitempty"`
				Long            string   `yaml:"long,omitempty"`
				Example         string   `yaml:"example,omitempty"`
				Deprecated      string   `yaml:"deprecated,omitempty"`
				Hidden          bool     `yaml:"hidden,omitempty"`
				CommandFlags    []*Flag  `yaml:"command_flags,omitempty"`
				PersistentFlags []*Flag  `yaml:"persistent_flags,omitempty"`
			}
			buildCommand := func(cmd *cobra.Command) *Command {
				doc := &Command{
					Name:       cmd.Name(),
					Path:       cmd.CommandPath(),
					Use:        strings.TrimSpace(strings.TrimPrefix(cmd.Use, cmd.Name())),
					Aliases:    cmd.Aliases,
					Short:      cmd.Short,
					Long:       cmd.Long,
					Example:    cmd.Example,
					Deprecated: cmd.Deprecated,
					Hidden:     cmd.Hidden,
				}
				if cmd.Parent() != nil {
					doc.ParentPath = cmd.Parent().CommandPath()
				}
				cmd.LocalNonPersistentFlags().VisitAll(func(flag *pflag.Flag) {
					doc.CommandFlags = append(doc.CommandFlags, buildFlag(flag))
				})
				cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
					doc.PersistentFlags = append(doc.PersistentFlags, buildFlag(flag))
				})
				return doc
			}

			out := make(map[string]*Command)
			var buildTree func(cmd *cobra.Command)
			buildTree = func(cmd *cobra.Command) {
				out[cmd.CommandPath()] = buildCommand(cmd)
				for _, sub := range cmd.Commands() {
					buildTree(sub)
				}
			}
			buildTree(root)

			dir, _ := cmd.Flags().GetString("out")
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return err
				}
			}

			b, err := yaml.Marshal(out)
			if err != nil {
				return err
			}

			return ioutil.WriteFile(filepath.Join(dir, root.Name()+".yml"), b, 0644)
		},
	}
	cmd.Flags().StringP("out", "o", "doc", "output directory")
	return cmd
}
