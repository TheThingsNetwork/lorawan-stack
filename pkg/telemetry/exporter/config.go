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

// Package telemetry contains the telemetry configuration and the hash generator that is used in other telemetry related
// packages.
package telemetry

import "time"

// CLI contains information regarding the telemetry collection for the CLI.
type CLI struct {
	Enable bool   `name:"enable" description:"Enables telemetry for CLI"`
	Target string `name:"target" description:"Target to which the information will be sent to"`
}

// EntityCountTelemetry contains information regarding the telemetry collection for the amount of entities.
type EntityCountTelemetry struct {
	Enable   bool          `name:"enable" description:"Enables entity count collection"`
	Interval time.Duration `name:"interval" description:"Interval between each run of the collection"`
}

// Config contains information regarding the telemetry collection.
type Config struct {
	Enable bool `name:"enable" description:"Enables telemetry collection"`
	// UIDElements is a list of elements that will be used to generate the UID.
	UIDElements          []string             `name:"-"`
	Target               string               `name:"target" description:"Target to which the information will be sent to"`                                 // nolint:lll
	NumConsumers         uint64               `name:"num-consumers" description:"Number of consumers that will be used to monitor telemetry related tasks"` // nolint:lll
	EntityCountTelemetry EntityCountTelemetry `name:"entity-count-telemetry"`
}
