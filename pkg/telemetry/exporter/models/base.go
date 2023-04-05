// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

// Package models contains telemetry models.
package models

// CLITelemetry contains telemetry information about the CLI.
type CLITelemetry struct{}

// OSTelemetry contains telemetry information about the operating system.
type OSTelemetry struct {
	OperatingSystem string `json:"operating_system"`
	Arch            string `json:"arch"`
	BinaryVersion   string `json:"binary_version" `
	GolangVersion   string `json:"golang_version"`
}
