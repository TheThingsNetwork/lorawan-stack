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
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
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
			dir, _ := cmd.Flags().GetString("out")
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return err
				}
			}
			disableAutoGenTag(root)
			return doc.GenYamlTree(root, dir)
		},
	}
	cmd.Flags().StringP("out", "o", "doc", "output directory")
	return cmd
}
