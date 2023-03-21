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

package alcsyncv1

import (
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Result is the result of a command execution.
type Result interface {
	// MarshalBinary marshals the result into a binary representation.
	MarshalBinary() ([]byte, error)

	// GetEvtSuccessfullyExecuted returns the event that should be emitted when the result is successful.
	GetEvtSuccessfullyExecuted() events.Builder
}

// Command is the interface for commands.
type Command interface {
	// Code returns the command code.
	Code() ttnpb.ALCSyncCommandIdentifier

	// GetEvtSuccessfullyParsed returns the event that should be emitted when the command is successfully parsed.
	GetEvtSuccessfullyParsed() events.Builder

	// Execute runs the command logic.
	Execute() (Result, error)
}
