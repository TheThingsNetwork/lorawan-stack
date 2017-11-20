// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"reflect"

	"github.com/TheThingsNetwork/ttn/pkg/validate"
)

// Validate is used as validator function by the GRPC validator interceptor.
func (req *CreateUserRequest) Validate() error {
	return validate.All(
		validate.Field(req.User.UserID, validate.ID).DescribeFieldName("User ID"),
		validate.Field(req.User.Password, validate.Password).DescribeFieldName("Password"),
		validate.Field(req.User.Email, validate.Email).DescribeFieldName("Email"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateUserRequest) Validate() error {
	mask := req.GetUpdateMask()
	if reflect.DeepEqual(mask, UserMask{}) {
		return ErrEmptyUpdateMask.New(nil)
	}

	validations := make([]validate.Errors, 0)

	if mask.Email {
		err := validate.Field(req.User.Email, validate.Email).DescribeFieldName("Email")
		if err != nil {
			validations = append(validations, err)
		}
	}

	return validate.All(validations...)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateUserPasswordRequest) Validate() error {
	return validate.All(
		validate.Field(req.Old, validate.Required).DescribeFieldName("Old password"),
		validate.Field(req.New, validate.Password).DescribeFieldName("New password"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *ValidateUserEmailRequest) Validate() error {
	return validate.Field(req.Token, validate.Required).DescribeFieldName("Token")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *CreateApplicationRequest) Validate() error {
	return validate.Field(req.Application.ApplicationID, validate.ID).DescribeFieldName("Application ID")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateApplicationRequest) Validate() error {
	mask := req.GetUpdateMask()
	if reflect.DeepEqual(mask, ApplicationMask{}) {
		return ErrEmptyUpdateMask.New(nil)
	}

	return nil
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *GenerateApplicationAPIKeyRequest) Validate() error {
	return validate.All(
		validate.Field(req.ApplicationID, validate.ID).DescribeFieldName("Application ID"),
		validate.Field(req.Name, validate.Required).DescribeFieldName("Key name"),
		validate.Field(req.Rights, validate.MinLength(1), validate.In(AllApplicationRights)).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateApplicationAPIKeyRequest) Validate() error {
	mask := req.GetUpdateMask()
	if reflect.DeepEqual(mask, APIKeyMask{}) {
		return ErrEmptyUpdateMask.New(nil)
	}

	validations := make([]validate.Errors, 0)

	err := validate.Field(req.ApplicationID, validate.ID).DescribeFieldName("Application ID")
	if err != nil {
		validations = append(validations, err)
	}

	if mask.Name {
		err := validate.Field(req.Key.Name, validate.Required).DescribeFieldName("Key name")
		if err != nil {
			validations = append(validations, err)
		}
	}

	if mask.Rights {
		err := validate.Field(req.Key.Rights, validate.MinLength(1), validate.In(AllApplicationRights)).DescribeFieldName("Key rights")
		if err != nil {
			validations = append(validations, err)
		}
	}

	return validate.All(validations...)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *RemoveApplicationAPIKeyRequest) Validate() error {
	return validate.All(
		validate.Field(req.ApplicationID, validate.ID).DescribeFieldName("Application ID"),
		validate.Field(req.Key, validate.Required).DescribeFieldName("API key"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *SetApplicationCollaboratorRequest) Validate() error {
	return validate.All(
		validate.Field(req.ApplicationID, validate.ID).DescribeFieldName("Application ID"),
		validate.Field(req.UserID, validate.ID).DescribeFieldName("User ID"),
		validate.Field(req.Rights, validate.NotRequired, validate.In(AllApplicationRights)).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *CreateGatewayRequest) Validate() error {
	return validate.All(
		validate.Field(req.Gateway.GatewayID, validate.ID).DescribeFieldName("Gateway ID"),
		validate.Field(req.Gateway.FrequencyPlanID, validate.Required).DescribeFieldName("Frequency plan ID"),
		validate.Field(req.Gateway.ClusterAddress, validate.Required).DescribeFieldName("Cluster adddress"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateGatewayRequest) Validate() error {
	mask := req.GetUpdateMask()
	if reflect.DeepEqual(mask, GatewayMask{}) {
		return ErrEmptyUpdateMask.New(nil)
	}

	validations := make([]validate.Errors, 0)

	err := validate.Field(req.Gateway.GatewayID, validate.ID).DescribeFieldName("Gateway ID")
	if err != nil {
		validations = append(validations, err)
	}

	if mask.FrequencyPlanID {
		err := validate.Field(req.Gateway.FrequencyPlanID, validate.Required).DescribeFieldName("Frequency plan ID")
		if err != nil {
			validations = append(validations, err)
		}
	}

	if mask.ClusterAddress {
		err := validate.Field(req.Gateway.ClusterAddress, validate.Required).DescribeFieldName("Cluster address")
		if err != nil {
			validations = append(validations, err)
		}
	}

	return validate.All(validations...)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *GenerateGatewayAPIKeyRequest) Validate() error {
	return validate.All(
		validate.Field(req.GatewayID, validate.ID).DescribeFieldName("Gateway ID"),
		validate.Field(req.Name, validate.Required).DescribeFieldName("Key name"),
		validate.Field(req.Rights, validate.MinLength(1), validate.In(AllGatewayRights)).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateGatewayAPIKeyRequest) Validate() error {
	mask := req.GetUpdateMask()
	if reflect.DeepEqual(mask, APIKeyMask{}) {
		return ErrEmptyUpdateMask.New(nil)
	}

	validations := make([]validate.Errors, 0)

	err := validate.Field(req.GatewayID, validate.ID).DescribeFieldName("Gateway ID")
	if err != nil {
		validations = append(validations, err)
	}

	if mask.Name {
		err := validate.Field(req.Key.Name, validate.Required).DescribeFieldName("Key name")
		if err != nil {
			validations = append(validations, err)
		}
	}

	if mask.Rights {
		err := validate.Field(req.Key.Rights, validate.MinLength(1), validate.In(AllGatewayRights)).DescribeFieldName("Key rights")
		if err != nil {
			validations = append(validations, err)
		}
	}

	return validate.All(validations...)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *RemoveGatewayAPIKeyRequest) Validate() error {
	return validate.All(
		validate.Field(req.GatewayID, validate.ID).DescribeFieldName("Gateway ID"),
		validate.Field(req.Key, validate.Required).DescribeFieldName("API key"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *SetGatewayCollaboratorRequest) Validate() error {
	return validate.All(
		validate.Field(req.GatewayID, validate.ID).DescribeFieldName("Gateway ID"),
		validate.Field(req.UserID, validate.ID).DescribeFieldName("User ID"),
		validate.Field(req.Rights, validate.NotRequired, validate.In(AllGatewayRights)).DescribeFieldName("Rights"),
	)
}

// validClientRights is the list of valid rights for a third-party client scope.
var validClientRights = []Right{
	RIGHT_USER_PROFILE_READ,
	RIGHT_USER_PROFILE_WRITE,
	RIGHT_USER_APPLICATIONS_LIST,
	RIGHT_USER_APPLICATIONS_CREATE,
	RIGHT_USER_APPLICATIONS,
	RIGHT_USER_GATEWAYS_LIST,
	RIGHT_USER_GATEWAYS_CREATE,
	RIGHT_USER_GATEWAYS,
	RIGHT_APPLICATION_INFO,
	RIGHT_APPLICATION_SETTINGS_BASIC,
	RIGHT_APPLICATION_SETTINGS_COLLABORATORS,
	RIGHT_APPLICATION_DELETE,
	RIGHT_APPLICATION_DEVICES_READ,
	RIGHT_APPLICATION_DEVICES_WRITE,
	RIGHT_APPLICATION_TRAFFIC_READ,
	RIGHT_APPLICATION_TRAFFIC_WRITE,
	RIGHT_GATEWAY_INFO,
	RIGHT_GATEWAY_SETTINGS_BASIC,
	RIGHT_GATEWAY_SETTINGS_COLLABORATORS,
	RIGHT_GATEWAY_DELETE,
	RIGHT_GATEWAY_TRAFFIC,
	RIGHT_GATEWAY_STATUS,
	RIGHT_GATEWAY_LOCATION,
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *CreateClientRequest) Validate() error {
	return validate.All(
		validate.Field(req.Client.ClientID, validate.ID).DescribeFieldName("Client ID"),
		validate.Field(req.Client.Description, validate.Required).DescribeFieldName("Description"),
		validate.Field(req.Client.RedirectURI, validate.Required).DescribeFieldName("Redirect URI"),
		validate.Field(req.Client.Rights, validate.MinLength(1), validate.In(validClientRights)).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateClientRequest) Validate() error {
	mask := req.GetUpdateMask()
	if reflect.DeepEqual(mask, ClientMask{}) {
		return ErrEmptyUpdateMask.New(nil)
	}

	validations := make([]validate.Errors, 0)

	err := validate.Field(req.Client.ClientID, validate.ID).DescribeFieldName("Client ID")
	if err != nil {
		validations = append(validations, err)
	}

	if mask.Description {
		err := validate.Field(req.Client.Description, validate.Required).DescribeFieldName("Description")
		if err != nil {
			validations = append(validations, err)
		}
	}

	if mask.RedirectURI {
		err := validate.Field(req.Client.RedirectURI, validate.Required).DescribeFieldName("Redirect URI")
		if err != nil {
			validations = append(validations, err)
		}
	}

	if mask.Rights {
		err := validate.Field(req.Client.Rights, validate.MinLength(1), validate.In(validClientRights)).DescribeFieldName("Rights")
		if err != nil {
			validations = append(validations, err)
		}
	}

	return validate.All(validations...)
}
