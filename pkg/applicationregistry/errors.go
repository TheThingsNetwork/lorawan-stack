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

package applicationregistry

import (
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
)

var (
	errApplicationNotFound = errors.DefineNotFound("application_not_found", "application not found")
	errTooManyApplications = errors.Define("too_many_applications", "too many applications found")
	errProcessorFailed     = errors.Define("processor_failed", "failed to process arguments")
	errNilApplication      = errors.DefineInvalidArgument("nil_application", "application specified is nil")
	errNilIdentifiers      = errors.DefineInvalidArgument("nil_identifiers", "identifiers specified are nil")
)
