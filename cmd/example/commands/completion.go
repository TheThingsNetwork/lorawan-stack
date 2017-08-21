// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCommand = &cobra.Command{
	Use:   "complete",
	Short: "Generate completions",
}

var completionBashCommand = &cobra.Command{
	Use:   "bash",
	Short: "Generate completion functions for bash",
	Run: func(*cobra.Command, []string) {
		err := Root.GenBashCompletion(os.Stdout)
		if err != nil {
			fmt.Println(err)
		}
	},
}

var completionZshCommand = &cobra.Command{
	Use:   "zsh",
	Short: "Generate completion functions for zsh",
	Run: func(*cobra.Command, []string) {
		err := Root.GenZshCompletion(os.Stdout)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	Root.AddCommand(completionCommand)
	completionCommand.AddCommand(completionBashCommand)
	completionCommand.AddCommand(completionZshCommand)
}
