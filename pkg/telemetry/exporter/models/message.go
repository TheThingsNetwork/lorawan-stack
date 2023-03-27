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

package models

// TelemetryMessage contains all the telemetry data that is to be exported by all services and CLI.
//
// This message is not supposed to be sent in its fullest. Meaning that tasks related to EntitiesCount should send only
// data regarding the entity amount, while the CLI information should be empty.
type TelemetryMessage struct {
	UID           string         `json:"uid"`
	OS            *OSTelemetry   `json:"os,omitempty"`
	CLI           *CLITelemetry  `json:"cli"`
	EntitiesCount *EntitiesCount `json:"entities_count"`
}
