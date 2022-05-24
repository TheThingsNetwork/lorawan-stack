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
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
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
				if err := os.MkdirAll(dir, 0o755); err != nil {
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

const MDDocFrontmatterTemplate = `---
title: "%s"
slug: %s
---

`

// GenMDDoc generates markdown documentation for the given root command.
func GenMDDoc(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "gen-md-doc",
		Hidden: true,
		Short:  fmt.Sprintf("Generate markdown documentation for %s", root.Name()),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("out")
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return err
				}
			}
			disableAutoGenTag(root)
			prepender := func(filename string) string {
				name := filepath.Base(filename)
				base := strings.TrimSuffix(name, path.Ext(name))
				title := strings.Replace(base, "_", " ", -1)
				fmt.Printf(`Write "%s" to %s`+"\n", title, filename)
				return fmt.Sprintf(MDDocFrontmatterTemplate, title, base)
			}

			linkHandler := func(name string) string {
				base := strings.TrimSuffix(name, path.Ext(name))
				return fmt.Sprintf(`{{< relref "%s" >}}`, strings.ToLower(base))
			}
			return doc.GenMarkdownTreeCustom(root, dir, prepender, linkHandler)
		},
	}
	cmd.Flags().StringP("out", "o", "doc", "output directory")
	return cmd
}

type command struct {
	Short       string             `json:"short,omitempty"`
	Path        string             `json:"path,omitempty"`
	SubCommands map[string]command `json:"subCommands,omitempty"`
}

func commandTree(cmd *cobra.Command) (res command) {
	res.Path = cmd.CommandPath()
	res.Short = cmd.Short
	if len(cmd.Commands()) == 0 {
		return
	}
	res.SubCommands = make(map[string]command, len(cmd.Commands()))
	for _, cmd := range cmd.Commands() {
		if !cmd.IsAvailableCommand() || cmd.IsAdditionalHelpTopicCommand() {
			continue
		}
		res.SubCommands[cmd.Name()] = commandTree(cmd)
	}
	return
}

// GenTree generates a JSON tree for the given root command
func GenJSONTree(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "gen-json-tree",
		Hidden: true,
		Short:  fmt.Sprintf("Generate JSON tree for %s", root.Name()),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("out")

			out := filepath.Join(dir, root.Name()+".json")

			f, err := os.Create(out)
			if err != nil {
				return err
			}
			defer f.Close()

			enc := json.NewEncoder(f)
			enc.SetIndent("", "  ")
			return enc.Encode(map[string]command{
				cmd.Root().Name(): commandTree(cmd.Root()),
			})
		},
	}
	cmd.Flags().StringP("out", "o", "doc", "output directory")
	return cmd
}
