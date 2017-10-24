// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/validate"
)

// Validate is used as validator function by the GRPC validator interceptor.
func (req *CreateUserRequest) Validate() error {
	return validate.All(
		validate.Field(req.UserID, validate.ID).DescribeFieldName("User ID"),
		validate.Field(req.Password, validate.Password).DescribeFieldName("Password"),
		validate.Field(req.Email, validate.Email).DescribeFieldName("Email"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateUserRequest) Validate() error {
	if req.GetUpdateMask() == nil {
		return ErrUpdateMaskNotFound.New(nil)
	}

	validations := make([]validate.Errors, 0)

	var err validate.Errors
	for _, path := range req.GetUpdateMask().Paths {
		switch path {
		case PathUserName:
		case PathUserEmail:
			err = validate.Field(req.Email, validate.Email).DescribeFieldName("Email")
		default:
			return ErrInvalidPathUpdateMask.New(errors.Attributes{
				"path": path,
			})
		}

		validations = append(validations, err)
	}

	return validate.All(validations...)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateUserPasswordRequest) Validate() error {
	return validate.Field(req.New, validate.Password).DescribeFieldName("Password")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *CreateApplicationRequest) Validate() error {
	return validate.Field(req.ApplicationID, validate.ID).DescribeFieldName("Application ID")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateApplicationRequest) Validate() error {
	if req.GetUpdateMask() == nil {
		return ErrUpdateMaskNotFound.New(nil)
	}

	validations := make([]validate.Errors, 0)

	if err := validate.Field(req.ApplicationID, validate.ID).DescribeFieldName("Application ID"); err != nil {
		validations = append(validations, err)
	}

	for _, path := range req.GetUpdateMask().Paths {
		switch path {
		case PathApplicationDescription:
		default:
			return ErrInvalidPathUpdateMask.New(errors.Attributes{
				"path": path,
			})
		}
	}

	return validate.All(validations...)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *GenerateApplicationAPIKeyRequest) Validate() error {
	return validate.All(
		validate.Field(req.KeyName, validate.Required).DescribeFieldName("Key name"),
		validate.Field(req.Rights, validate.Required, validate.MinLength(1)).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *RemoveApplicationAPIKeyRequest) Validate() error {
	return validate.Field(req.KeyName, validate.Required).DescribeFieldName("Key name")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *SetApplicationCollaboratorRequest) Validate() error {
	return validate.All(
		validate.Field(req.ApplicationID, validate.ID).DescribeFieldName("Application ID"),
		validate.Field(req.UserID, validate.ID).DescribeFieldName("User ID"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *CreateGatewayRequest) Validate() error {
	return validate.All(
		validate.Field(req.GatewayID, validate.ID).DescribeFieldName("Gateway ID"),
		validate.Field(req.FrequencyPlanID, validate.Required).DescribeFieldName("Frequency plan ID"),
		validate.Field(req.ClusterAddress, validate.Required).DescribeFieldName("Cluster adddress"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateGatewayRequest) Validate() error {
	if req.GetUpdateMask() == nil {
		return ErrUpdateMaskNotFound.New(nil)
	}

	validations := make([]validate.Errors, 0)

	if err := validate.Field(req.GatewayID, validate.ID).DescribeFieldName("Gateway ID"); err != nil {
		validations = append(validations, err)
	}

	var err validate.Errors
	for _, path := range req.GetUpdateMask().Paths {
		switch path {
		case PathGatewayDescription,
			PathGatewayPrivacySettingsStatusPublic,
			PathGatewayPrivacySettingsLocationPublic,
			PathGatewayPrivacySettingsContactable,
			PathGatewayAutoUpdate,
			PathGatewayPlatform,
			PathGatewayAntennas,
			PathGatewayAttributes,
			PathGatewayContactAccount:
		case PathGatewayFrequencyPlanID:
			err = validate.Field(req.FrequencyPlanID, validate.Required).DescribeFieldName("Frequency plan ID")
		case PathGatewayClusterAddress:
			err = validate.Field(req.ClusterAddress, validate.Required).DescribeFieldName("Cluster address")
		default:
			return ErrInvalidPathUpdateMask.New(errors.Attributes{
				"path": path,
			})
		}

		validations = append(validations, err)
	}

	return validate.All(validations...)

}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *SetGatewayCollaboratorRequest) Validate() error {
	return validate.All(
		validate.Field(req.GatewayID, validate.ID).DescribeFieldName("Gateway ID"),
		validate.Field(req.UserID, validate.ID).DescribeFieldName("User ID"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *CreateClientRequest) Validate() error {
	return validate.All(
		validate.Field(req.ClientID, validate.ID).DescribeFieldName("Client ID"),
		validate.Field(req.RedirectURI, validate.Required).DescribeFieldName("Callback URI"),
		validate.Field(req.Grants, validate.MinLength(1)).DescribeFieldName("Grants"),
		validate.Field(req.Rights, validate.MinLength(1)).DescribeFieldName("Scope"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateClientRequest) Validate() error {
	if req.GetUpdateMask() == nil {
		return ErrUpdateMaskNotFound.New(nil)
	}

	validations := make([]validate.Errors, 0)

	if err := validate.Field(req.ClientID, validate.ID).DescribeFieldName("Client ID"); err != nil {
		validations = append(validations, err)
	}

	var err validate.Errors
	for _, path := range req.GetUpdateMask().Paths {
		switch path {
		case PathClientDescription:
		case PathClientCallbackURI:
			err = validate.Field(req.RedirectURI, validate.Required).DescribeFieldName("Callback URI")
		case PathClientGrants:
			err = validate.Field(req.Grants, validate.MinLength(1)).DescribeFieldName("Grants")
		case PathClientScope:
			err = validate.Field(req.Rights, validate.MinLength(1)).DescribeFieldName("Scope")
		default:
			return ErrInvalidPathUpdateMask.New(errors.Attributes{
				"path": path,
			})
		}

		validations = append(validations, err)
	}

	return validate.All(validations...)

}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *SetClientStateRequest) Validate() error {
	allStates := []ClientState{STATE_PENDING, STATE_REJECTED, STATE_APPROVED}

	return validate.All(
		validate.Field(req.ClientID, validate.ID).DescribeFieldName("Client ID"),
		validate.Field(req.State, validate.In(allStates)).DescribeFieldName("State"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *SetClientCollaboratorRequest) Validate() error {
	return validate.All(
		validate.Field(req.ClientID, validate.ID).DescribeFieldName("Client ID"),
		validate.Field(req.UserID, validate.ID).DescribeFieldName("User ID"),
	)
}
