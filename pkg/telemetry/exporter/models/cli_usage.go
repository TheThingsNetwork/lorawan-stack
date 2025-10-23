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

package models

// CommandUsageData holds aggregated data for a single command's usage.
type CommandUsageData struct {
	CommandPath    string `json:"command_path"`
	ExecutionCount int    `json:"execution_count"`
	LastUsedAt     int64  `json:"last_used_at"`
}

// AliasUsageData holds aggregated data for a single alias's usage.
type AliasUsageData struct {
	CommandName string `json:"command_name"`
	Alias       string `json:"alias"`
	UsageCount  int    `json:"usage_count"`
}

// CmdData represents a snapshot of the aggregated telemetry data.
type CmdData struct {
	Commands []CommandUsageData `json:"commands"`
	Aliases  []AliasUsageData   `json:"aliases"`
}
