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
