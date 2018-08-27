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
func (ids UserIdentifiers) Validate() error {
	if ids.IsZero() {
		return errMissingIdentifiers
	}

	return validate.All(
		validate.Field(ids.UserID, validate.NotRequired, validate.ID).DescribeFieldName("User ID"),
		validate.Field(ids.Email, validate.NotRequired, validate.Email).DescribeFieldName("Email"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (ids ApplicationIdentifiers) Validate() error {
	return validate.Field(ids.ApplicationID, validate.ID).DescribeFieldName("Application ID")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (ids GatewayIdentifiers) Validate() error {
	if ids.IsZero() {
		return errMissingIdentifiers
	}

	return validate.All(
		validate.Field(ids.GatewayID, validate.NotRequired, validate.ID).DescribeFieldName("Gateway ID"),
		validate.Field(ids.EUI, validate.NotRequired).DescribeFieldName("EUI"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (ids ClientIdentifiers) Validate() error {
	return validate.Field(ids.ClientID, validate.ID).DescribeFieldName("Client ID")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (ids OrganizationIdentifiers) Validate() error {
	return validate.Field(ids.OrganizationID, validate.ID).DescribeFieldName("Organization ID")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (ids OrganizationOrUserIdentifiers) Validate() error {
	if ids := ids.GetUserIDs(); ids != nil {
		return ids.Validate()
	}

	if ids := ids.GetOrganizationIDs(); ids != nil {
		return ids.Validate()
	}

	return errMissingIdentifiers
}

// Validate is used as validator function by the GRPC validator interceptor.
func (ids *EndDeviceIdentifiers) Validate() error {
	if ids == nil || ids.IsZero() {
		return errMissingIdentifiers
	}
	return validate.All(
		validate.Field(ids.GetDeviceID(), validate.NotRequired, validate.ID).DescribeFieldName("Device ID"),
		validate.Field(ids.GetApplicationID(), validate.NotRequired, validate.ID).DescribeFieldName("Application ID"),
	)
}
