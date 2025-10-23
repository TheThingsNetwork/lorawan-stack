// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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

package store

import (
	"os"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter/models"
)

// UsageData handles the store interaction regarding commands used.
type UsageData interface {
	// StoreCommand stores a successful command execution.
	StoreCommand(cmd *cobra.Command) error
	// GetCmdData retrieves the current aggregated telemetry data.
	GetCmdData() (*models.CmdData, error)
	// ResetCmdData clears all the aggregated data.
	ResetCmdData() error
}

func (st *dbStore) StoreCommand(cmd *cobra.Command) error {
	data, err := st.Read()
	if err != nil {
		return err
	}

	// Store command usage
	commandPath := strings.Join(strings.Split(cmd.CommandPath(), " ")[1:], " ")
	commandFound := false
	for idx := range data.Cmd.Commands {
		if data.Cmd.Commands[idx].CommandPath == commandPath {
			data.Cmd.Commands[idx].ExecutionCount = data.Cmd.Commands[idx].ExecutionCount + 1
			data.Cmd.Commands[idx].LastUsedAt = time.Now().Unix()
			commandFound = true
			break
		}
	}
	if !commandFound {
		data.Cmd.Commands = append(data.Cmd.Commands, models.CommandUsageData{
			CommandPath:    commandPath,
			ExecutionCount: 1,
			LastUsedAt:     time.Now().Unix(),
		})
	}

	// Store alias usage
	aliases := getAliases(cmd)
	for name, alias := range aliases {
		found := false
		for idx := range data.Cmd.Aliases {
			if data.Cmd.Aliases[idx].CommandName == name && data.Cmd.Aliases[idx].Alias == alias {
				data.Cmd.Aliases[idx].UsageCount = data.Cmd.Aliases[idx].UsageCount + 1
				found = true
				break
			}
		}
		if !found {
			data.Cmd.Aliases = append(data.Cmd.Aliases, models.AliasUsageData{
				Alias:       alias,
				CommandName: name,
				UsageCount:  1,
			})
		}
	}

	return st.Write(data)
}

func (st *dbStore) GetCmdData() (*models.CmdData, error) {
	data, err := st.Read()
	if err != nil {
		return nil, err
	}
	return &data.Cmd, nil
}

func (st *dbStore) ResetCmdData() error {
	data, err := st.Read()
	if err != nil {
		return err
	}

	data.Cmd.Aliases = make([]models.AliasUsageData, 0)
	data.Cmd.Commands = make([]models.CommandUsageData, 0)

	return st.Write(data)
}

// getAliases extracts the aliases used for the given command and its parents.
//
// If a user runs ttn-lw-cli u c (where u is an alias for users and c is an alias for create), the getAliases function
// would return a map like this:
//
//	{
//	  "users": "u",
//	  "create": "c"
//	}
func getAliases(cmd *cobra.Command) map[string]string {
	aliases := make(map[string]string)
	args := os.Args
	// Iterate from the executed command up to the root.
	for c := cmd; c != nil && c != cmd.Root(); c = c.Parent() {
		// Search backwards through the remaining command-line arguments.
		for i := len(args) - 1; i >= 0; i-- {
			// Check if the argument is a known alias for the current command.
			if slices.Contains(c.Aliases, args[i]) {
				aliases[c.Name()] = args[i]
				// Remove the found alias and subsequent arguments from consideration for parent commands.
				args = args[:i]
				break
			}
		}
	}
	return aliases
}
