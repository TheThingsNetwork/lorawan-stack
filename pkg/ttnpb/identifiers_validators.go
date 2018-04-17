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

import "github.com/TheThingsNetwork/ttn/pkg/validate"

// Validate is used as validator function by the GRPC validator interceptor.
func (i UserIdentifiers) Validate() error {
	if i.IsZero() {
		return ErrEmptyIdentifiers.New(nil)
	}

	return validate.All(
		validate.Field(i.UserID, validate.NotRequired, validate.ID).DescribeFieldName("User ID"),
		validate.Field(i.Email, validate.NotRequired, validate.Email).DescribeFieldName("Email"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (i ApplicationIdentifiers) Validate() error {
	return validate.Field(i.ApplicationID, validate.ID).DescribeFieldName("Application ID")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (i GatewayIdentifiers) Validate() error {
	if i.IsZero() {
		return ErrEmptyIdentifiers.New(nil)
	}

	return validate.All(
		validate.Field(i.GatewayID, validate.NotRequired, validate.ID).DescribeFieldName("Gateway ID"),
		validate.Field(i.EUI, validate.NotRequired).DescribeFieldName("EUI"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (i ClientIdentifiers) Validate() error {
	return validate.Field(i.ClientID, validate.ID).DescribeFieldName("Client ID")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (i OrganizationIdentifiers) Validate() error {
	return validate.Field(i.OrganizationID, validate.ID).DescribeFieldName("Organization ID")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (i OrganizationOrUserIdentifiers) Validate() error {
	if id := i.GetUserID(); id != nil {
		return id.Validate()
	}

	if id := i.GetOrganizationID(); id != nil {
		return id.Validate()
	}

	return ErrEmptyIdentifiers.New(nil)
}
