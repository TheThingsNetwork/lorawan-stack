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

import "go.thethings.network/lorawan-stack/pkg/validate"

// Validate is used as validator function by the GRPC validator interceptor.
func (ids *ApplicationWebhookIdentifiers) Validate() error {
	if ids == nil || ids.IsZero() {
		return errMissingIdentifiers
	}
	return validate.All(
		validate.Field(ids.GetWebhookID(), validate.NotRequired, validate.ID).DescribeFieldName("Webhook ID"),
		validate.Field(ids.GetApplicationID(), validate.NotRequired, validate.ID).DescribeFieldName("Application ID"),
	)
}
