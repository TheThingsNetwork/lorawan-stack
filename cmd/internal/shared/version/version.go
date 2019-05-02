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

package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/pkg/version"
)

func print(k, v string) {
	fmt.Printf("%-20s %s\n", k+":", v)
}

// Print version information.
func Print(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s: %s\n", root.Short, root.Name())
			print("Version", version.TTN)
			if version.BuildDate != "" {
				print("Build date", version.BuildDate)
			}
			if version.GitCommit != "" {
				print("Git commit", version.GitCommit)
			}
			print("Go version", runtime.Version())
			print("OS/Arch", runtime.GOOS+"/"+runtime.GOARCH)
		},
	}
}
