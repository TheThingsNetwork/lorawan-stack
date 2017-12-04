// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCommand = &cobra.Command{
	Use:   "completion",
	Short: "Generate completions",
}

var completionBashCommand = &cobra.Command{
	Use:   "bash",
	Short: "Generate completion functions for bash",
	RunE: func(*cobra.Command, []string) error {
		return Root.GenBashCompletion(os.Stdout)
	},
}

var completionZshCommand = &cobra.Command{
	Use:   "zsh",
	Short: "Generate completion functions for zsh",
	RunE: func(*cobra.Command, []string) error {
		return Root.GenZshCompletion(os.Stdout)
	},
}

func init() {
	Root.AddCommand(completionCommand)
	completionCommand.AddCommand(completionBashCommand)
	completionCommand.AddCommand(completionZshCommand)
}
