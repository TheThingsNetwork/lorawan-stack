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
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

var errInvalidShell = errors.DefineInvalidArgument("invalid_shell", "invalid shell")

// Complete returns the auto-complete command
func Complete() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "complete",
		Hidden: true,
		Short:  "Generate script for auto-completion",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name, _ := cmd.Flags().GetString("executable"); name != "" {
				cmd.Root().Use = name
			}
			switch shell, _ := cmd.Flags().GetString("shell"); shell {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				// Fish does not accept `-` in variable names
				buf := new(bytes.Buffer)
				if err := cmd.Root().GenFishCompletion(buf, true); err != nil {
					return err
				}
				script := strings.Replace(buf.String(), "__ttn-lw-", "__ttn_lw_", -1)
				_, err := fmt.Print(script)
				return err
			case "powershell":
				return cmd.Root().GenPowerShellCompletion(os.Stdout)
			default:
				return errInvalidShell.WithAttributes("shell", shell)
			}
		},
	}
	cmd.Flags().String("shell", "bash", "bash|zsh|fish|powershell")
	cmd.Flags().String("executable", "", "Executable name to create generate auto completion script for")
	return cmd
}
