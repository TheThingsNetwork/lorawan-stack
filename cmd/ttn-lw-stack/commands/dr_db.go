// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"github.com/spf13/cobra"
)

var (
	drDBCommand = &cobra.Command{
		Use:   "dr-db",
		Short: "Device Repository commands",
	}
	drInitCommand = &cobra.Command{
		Use:   "init",
		Short: "Fetch Device Repository files and generate index",
		RunE: func(cmd *cobra.Command, args []string) error {
			overwrite, _ := cmd.Flags().GetBool("overwrite")

			return config.DR.Initialize(ctx, config.Blob, overwrite)
		},
	}
)

func init() {
	Root.AddCommand(drDBCommand)

	drInitCommand.Flags().Bool("overwrite", true, "Overwrite existing index files")
	drDBCommand.AddCommand(drInitCommand)
}
