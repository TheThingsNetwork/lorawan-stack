// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import "google.golang.org/protobuf/types/known/timestamppb"

// EmailValidationOverwrite is a function to overwrite the EmailValidation.
type EmailValidationOverwrite func(*EmailValidation) *EmailValidation

// EmailValidationWithExpiresAt overwrites the EmailValidation ExpiresAt field.
func EmailValidationWithExpiresAt(expiresAt *timestamppb.Timestamp) EmailValidationOverwrite {
	return func(validation *EmailValidation) *EmailValidation {
		validation.ExpiresAt = expiresAt
		return validation
	}
}
