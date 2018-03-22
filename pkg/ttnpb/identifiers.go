// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import "github.com/TheThingsNetwork/ttn/pkg/validate"

// Validate is used as validator function by the GRPC validator interceptor.
func (i UserIdentifiers) Validate() error {
	return validate.Field(i.UserID, validate.ID).DescribeFieldName("User ID")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (i ApplicationIdentifiers) Validate() error {
	return validate.Field(i.ApplicationID, validate.ID).DescribeFieldName("Application ID")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (i GatewayIdentifiers) Validate() error {
	return validate.Field(i.GatewayID, validate.ID).DescribeFieldName("Gateway ID")
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
