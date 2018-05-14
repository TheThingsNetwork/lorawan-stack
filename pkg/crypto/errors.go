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

package crypto

import (
	"fmt"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

var (
	errInvalidSize = func(typeName, expectedSize string, code errors.Code) *errors.ErrDescriptor {
		return &errors.ErrDescriptor{
			MessageFormat:  fmt.Sprintf("Expected %s to be %s, got {size}", typeName, expectedSize),
			SafeAttributes: []string{"size"},
			Code:           code,
			Type:           errors.InvalidArgument,
		}
	}

	// ErrInvalidJoinAcceptPayloadForEncryption is returned if encryption couldn't be completed due to
	// an invalid join-accept payload size.
	ErrInvalidJoinAcceptPayloadForEncryption = errInvalidSize("Join-accept payload", "16 or 32 bytes", 1)
	// ErrInvalidJoinAcceptPayloadForDecryption is returned if decryption couldn't be completed due to
	// an invalid encrypted join-accept payload size.
	ErrInvalidJoinAcceptPayloadForDecryption = errInvalidSize("Encrypted join-accept payload", "16 or 32 bytes", 2)
	// ErrInvalidJoinRequestPayloadForMIC is returned if MIC computing couldn't be completed due to
	// an invalid join-request payload size.
	ErrInvalidJoinRequestPayloadForMIC = errInvalidSize("Join-request payload", "19 bytes", 3)
	// ErrInvalidJoinAcceptPayloadForMIC is returned if MIC computing couldn't be completed due to
	// an invalid join-accept payload size.
	ErrInvalidJoinAcceptPayloadForMIC = errInvalidSize("Join-accept payload", "13 or 29 bytes", 4)

	// ErrInvalidRejoinRequestSizeForMIC is returned if MIC computing couldn't be completed due to
	// an invalid rejoin-request payload size.
	ErrInvalidRejoinRequestSizeForMIC = errInvalidSize("Rejoin-request payload", "at least 3 bytes", 5)
	// ErrInvalidRejoinRequestType0_2ForMIC is returned if MIC computing couldn't be completed due to
	// an invalid type 0 or 2 rejoin-request payload size.
	ErrInvalidRejoinRequestType0_2ForMIC = errInvalidSize("Rejoin-request type 0 or 2 payload", "15 bytes", 6)
	// ErrInvalidRejoinRequestType1ForMIC is returned if MIC computing couldn't be completed due to
	// an invalid type 1 rejoin-request payload size.
	ErrInvalidRejoinRequestType1ForMIC = errInvalidSize("Rejoin-request type 1 payload", "20 bytes", 7)

	// ErrNoKeyPresent is returned if no key could be read during a WrapKey operation.
	ErrNoKeyPresent = &errors.ErrDescriptor{
		MessageFormat: "No key present",
		Code:          8,
		Type:          errors.InvalidArgument,
	}
	// ErrInvalidPlaintextLength is returned when the plain text message doesn't have
	// the expected length.
	ErrInvalidPlaintextLength = errInvalidSize("Plaintext length", "a multiple of 8 bytes", 9)
	// ErrInvalidCiphertextLength is returned when the cipher text message doesn't have
	// the expected length.
	ErrInvalidCiphertextLength = errInvalidSize("Ciphertext length", "a multiple of 8 bytes", 10)
	// ErrCorruptKeyData is returned if the key data is corrupt.
	ErrCorruptKeyData = &errors.ErrDescriptor{
		MessageFormat: "Corrupt key data",
		Code:          11,
		Type:          errors.InvalidArgument,
	}
)

func init() {
	ErrInvalidJoinAcceptPayloadForEncryption.Register()
	ErrInvalidJoinAcceptPayloadForDecryption.Register()
	ErrInvalidJoinRequestPayloadForMIC.Register()
	ErrInvalidJoinAcceptPayloadForMIC.Register()
	ErrInvalidRejoinRequestSizeForMIC.Register()
	ErrInvalidRejoinRequestType0_2ForMIC.Register()
	ErrInvalidRejoinRequestType1ForMIC.Register()
	ErrNoKeyPresent.Register()
	ErrInvalidPlaintextLength.Register()
	ErrInvalidCiphertextLength.Register()
	ErrCorruptKeyData.Register()
}
