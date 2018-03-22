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

package ttnpb

import (
	"fmt"
	"math"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
)

var (
	// ErrEmptyUpdateMask is returned when the update mask is specified but empty.
	ErrEmptyUpdateMask = &errors.ErrDescriptor{
		MessageFormat: "update_mask must be non-empty",
		Code:          1,
		Type:          errors.InvalidArgument,
	}

	// ErrInvalidPathUpdateMask is returned when the update mask includes a wrong field path.
	ErrInvalidPathUpdateMask = &errors.ErrDescriptor{
		MessageFormat: "Invalid update_mask: `{path}` is not a valid path",
		Code:          2,
		Type:          errors.InvalidArgument,
	}

	// ErrMissingPayload represents error ocurring when message payload is missing.
	ErrMissingPayload = &errors.ErrDescriptor{
		MessageFormat: "Message payload is missing",
		Type:          errors.InvalidArgument,
		Code:          3,
	}

	// ErrMissingRawPayload represents error ocurring when raw message payload is missing.
	ErrMissingRawPayload = &errors.ErrDescriptor{
		MessageFormat: "Raw Message payload is missing",
		Type:          errors.InvalidArgument,
		Code:          4,
	}

	// ErrMissingDevAddr represents error ocurring when DevAddr is missing.
	ErrMissingDevAddr = &errors.ErrDescriptor{
		MessageFormat: "DevAddr is missing",
		Type:          errors.InvalidArgument,
		Code:          5,
	}

	// ErrWrongPayloadType represents error ocurring when wrong payload type is received.
	ErrWrongPayloadType = &errors.ErrDescriptor{
		MessageFormat:  "Wrong payload type: `{type}`",
		Type:           errors.InvalidArgument,
		Code:           6,
		SafeAttributes: []string{"type"},
	}

	// ErrMissingDevEUI represents error ocurring when DevEUI is missing.
	ErrMissingDevEUI = &errors.ErrDescriptor{
		MessageFormat: "DevEUI is missing",
		Type:          errors.InvalidArgument,
		Code:          7,
	}

	// ErrMissingJoinEUI represents error ocurring when JoinEUI is missing.
	ErrMissingJoinEUI = &errors.ErrDescriptor{
		MessageFormat: "JoinEUI is missing",
		Type:          errors.InvalidArgument,
		Code:          8,
	}

	// ErrFCntTooHigh represents error ocurring when FCnt provided is too high.
	ErrFCntTooHigh = &errors.ErrDescriptor{
		MessageFormat: fmt.Sprintf("FCnt must be lower or equal to %d", math.MaxUint16),
		Type:          errors.InvalidArgument,
		Code:          9,
	}

	// ErrFPortTooHigh represents error ocurring when FPort provided is too high.
	ErrFPortTooHigh = &errors.ErrDescriptor{
		MessageFormat: fmt.Sprintf("FPort must be lower or equal to %d", math.MaxUint8),
		Type:          errors.InvalidArgument,
		Code:          10,
	}

	// ErrTxChIdxTooHigh represents error ocurring when TxChIdx provided is too high.
	ErrTxChIdxTooHigh = &errors.ErrDescriptor{
		MessageFormat: fmt.Sprintf("TxChIdx must be lower or equal to %d", math.MaxUint8),
		Type:          errors.InvalidArgument,
		Code:          11,
	}

	// ErrTxDRIdxTooHigh represents error ocurring when TxDRIdx provided is too high.
	ErrTxDRIdxTooHigh = &errors.ErrDescriptor{
		MessageFormat: fmt.Sprintf("TxDRIdx must be lower or equal to %d", math.MaxUint8),
		Type:          errors.InvalidArgument,
		Code:          12,
	}

	// ErrEmptyDeviceIdentifiers represents error ocurring when EndDeviceIdentifiers provided are empty.
	ErrEmptyDeviceIdentifiers = &errors.ErrDescriptor{
		MessageFormat: "EndDeviceIdentifiers must be non-empty",
		Type:          errors.InvalidArgument,
		Code:          13,
	}

	// ErrMissingApplicationID represents error ocurring when ApplicationID is missing.
	ErrMissingApplicationID = &errors.ErrDescriptor{
		MessageFormat: "ApplicationID is missing",
		Type:          errors.InvalidArgument,
		Code:          14,
	}

	// ErrEmptyIdentifiers is returned when the XXXIdentifiers are empty.
	ErrEmptyIdentifiers = &errors.ErrDescriptor{
		MessageFormat: "Identifiers must be non-empty",
		Code:          15,
		Type:          errors.InvalidArgument,
	}
)

func init() {
	ErrEmptyUpdateMask.Register()
	ErrInvalidPathUpdateMask.Register()
	ErrMissingPayload.Register()
	ErrMissingRawPayload.Register()
	ErrMissingDevAddr.Register()
	ErrWrongPayloadType.Register()
	ErrMissingDevEUI.Register()
	ErrMissingJoinEUI.Register()
	ErrFCntTooHigh.Register()
	ErrFPortTooHigh.Register()
	ErrTxChIdxTooHigh.Register()
	ErrTxDRIdxTooHigh.Register()
	ErrWrongPayloadType.Register()
	ErrEmptyDeviceIdentifiers.Register()
	ErrMissingApplicationID.Register()
	ErrEmptyIdentifiers.Register()
}
