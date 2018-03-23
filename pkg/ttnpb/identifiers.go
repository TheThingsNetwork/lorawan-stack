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
	return validate.Field(i.UserID, validate.ID).DescribeFieldName("User ID")
}

// IsZero returns true if all identifiers have zero-values.
func (i UserIdentifiers) IsZero() bool {
	return i.UserID == ""
}

// Equals returns true if the receiver identifiers matches to other identifiers.
func (i UserIdentifiers) Equals(other UserIdentifiers) bool {
	return i.UserID == other.UserID
}

// Contains returns true if other is contained in the receiver.
func (i UserIdentifiers) Contains(other UserIdentifiers) bool {
	return i.UserID == other.UserID
}

// Validate is used as validator function by the GRPC validator interceptor.
func (i ApplicationIdentifiers) Validate() error {
	return validate.Field(i.ApplicationID, validate.ID).DescribeFieldName("Application ID")
}

// Contains returns true if other is contained in the receiver.
func (i ApplicationIdentifiers) Contains(other ApplicationIdentifiers) bool {
	return i.ApplicationID == other.ApplicationID
}

// IsZero returns true if all identifiers have zero-values.
func (i ApplicationIdentifiers) IsZero() bool {
	return i.ApplicationID == ""
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

// IsZero returns true if all identifiers have zero-values.
func (i GatewayIdentifiers) IsZero() bool {
	return i.GatewayID == "" && i.EUI == nil
}

// Contains returns true if other is contained in the receiver.
func (i GatewayIdentifiers) Contains(other GatewayIdentifiers) bool {
	if other.IsZero() {
		return i.IsZero()
	}

	return (other.GatewayID == "" || (other.GatewayID != "" && i.GatewayID == other.GatewayID)) &&
		((i.EUI == nil && other.EUI == nil) || (i.EUI != nil && other.EUI == nil) || i.EUI.Equal(*other.EUI))
}

// Validate is used as validator function by the GRPC validator interceptor.
func (i ClientIdentifiers) Validate() error {
	return validate.Field(i.ClientID, validate.ID).DescribeFieldName("Client ID")
}

// IsZero returns true if all identifiers have zero-values.
func (i ClientIdentifiers) IsZero() bool {
	return i.ClientID == ""
}

// Validate is used as validator function by the GRPC validator interceptor.
func (i OrganizationIdentifiers) Validate() error {
	return validate.Field(i.OrganizationID, validate.ID).DescribeFieldName("Organization ID")
}

// IsZero returns true if all identifiers have zero-values.
func (i OrganizationIdentifiers) IsZero() bool {
	return i.OrganizationID == ""
}

// Contains returns true if other is contained in the receiver.
func (i OrganizationIdentifiers) Contains(other OrganizationIdentifiers) bool {
	return i.OrganizationID == other.OrganizationID
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
