// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/validate"
)

// IsExpired checks whether or not the invitation is expired.
func (i *ListInvitationsResponse_Invitation) IsExpired() bool {
	return i.ExpiresAt.Before(time.Now().UTC())
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateSettingsRequest) Validate() error {
	mask := req.GetUpdateMask()
	paths := mask.GetPaths()

	if paths == nil || len(paths) == 0 {
		return ErrEmptyUpdateMask.New(nil)
	}

	validations := make([]validate.Errors, 0)

	var err validate.Errors
	for _, path := range paths {
		switch true {
		case FieldPathSettingsBlacklistedIDs.MatchString(path):
			for _, id := range req.Settings.BlacklistedIDs {
				err := validate.Field(id, validate.ID).DescribeFieldName("Blacklisted ID")
				if err != nil {
					validations = append(validations, err)
				}
			}
		case FieldPathSettingsUserRegistrationSkipValidation.MatchString(path),
			FieldPathSettingsUserRegistrationSelfRegistration.MatchString(path),
			FieldPathSettingsUserRegistrationAdminApproval.MatchString(path),
			FieldPathSettingsValidationTokenTTL.MatchString(path),
			FieldPathSettingsAllowedEmails.MatchString(path),
			FieldPathSettingsInvitationTokenTTL.MatchString(path):
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
	paths := mask.GetPaths()

	if paths == nil || len(paths) == 0 {
		return ErrEmptyUpdateMask.New(nil)
	}

	validations := make([]validate.Errors, 0)

	if err := validate.Field(req.User.UserID, validate.ID).DescribeFieldName("User ID"); err != nil {
		validations = append(validations, err)
	}

	var err validate.Errors
	for _, path := range paths {
		switch true {
		case FieldPathUserName.MatchString(path):
		case FieldPathUserEmail.MatchString(path):
			err = validate.Field(req.User.Email, validate.Email).DescribeFieldName("Email")
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
	return validate.All(
		validate.Field(req.Old, validate.Required).DescribeFieldName("Old password"),
		validate.Field(req.New, validate.Password).DescribeFieldName("New password"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *GenerateUserAPIKeyRequest) Validate() error {
	return validate.All(
		validate.Field(req.Name, validate.Required).DescribeFieldName("Key name"),
		validate.Field(req.Rights, validate.MinLength(1), validate.In(AllUserRights())).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateUserAPIKeyRequest) Validate() error {
	return validate.All(
		validate.Field(req.Name, validate.Required).DescribeFieldName("Key name"),
		validate.Field(req.Rights, validate.MinLength(1), validate.In(AllUserRights())).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *RemoveUserAPIKeyRequest) Validate() error {
	return validate.Field(req.Name, validate.Required).DescribeFieldName("Key name")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *ValidateUserEmailRequest) Validate() error {
	return validate.Field(req.Token, validate.Required).DescribeFieldName("Token")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *CreateApplicationRequest) Validate() error {
	return validate.All(
		validate.Field(req.Application.ApplicationID, validate.ID).DescribeFieldName("Application ID"),
		validate.Field(req.OrganizationID, validate.NotRequired, validate.ID).DescribeFieldName("Organization ID"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateApplicationRequest) Validate() error {
	mask := req.GetUpdateMask()
	paths := mask.GetPaths()

	if paths == nil || len(paths) == 0 {
		return ErrEmptyUpdateMask.New(nil)
	}

	validations := make([]validate.Errors, 0)

	if err := validate.Field(req.Application.ApplicationID, validate.ID).DescribeFieldName("Application ID"); err != nil {
		validations = append(validations, err)
	}

	for _, path := range paths {
		switch true {
		case FieldPathApplicationDescription.MatchString(path):
		default:
			return ErrInvalidPathUpdateMask.New(errors.Attributes{
				"path": path,
			})
		}
	}

	return validate.All(validations...)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *ListApplicationsRequest) Validate() error {
	return validate.Field(req.OrganizationID, validate.NotRequired, validate.ID).DescribeFieldName("Organization ID")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *GenerateApplicationAPIKeyRequest) Validate() error {
	return validate.All(
		validate.Field(req.ApplicationID, validate.ID).DescribeFieldName("Application ID"),
		validate.Field(req.Name, validate.Required).DescribeFieldName("Key name"),
		validate.Field(req.Rights, validate.MinLength(1), validate.In(AllApplicationRights())).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateApplicationAPIKeyRequest) Validate() error {
	return validate.All(
		validate.Field(req.ApplicationID, validate.ID).DescribeFieldName("Application ID"),
		validate.Field(req.Name, validate.Required).DescribeFieldName("Key name"),
		validate.Field(req.Rights, validate.MinLength(1), validate.In(AllApplicationRights())).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *RemoveApplicationAPIKeyRequest) Validate() error {
	return validate.All(
		validate.Field(req.ApplicationID, validate.ID).DescribeFieldName("Application ID"),
		validate.Field(req.Name, validate.Required).DescribeFieldName("Key name"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *ApplicationCollaborator) Validate() error {
	id := req.OrganizationOrUserIdentifier.GetUserID()
	if id == "" {
		id = req.OrganizationOrUserIdentifier.GetOrganizationID()
	}

	return validate.All(
		validate.Field(req.ApplicationID, validate.ID).DescribeFieldName("Application ID"),
		validate.Field(id, validate.ID).DescribeFieldName("Account ID"),
		validate.Field(req.Rights, validate.NotRequired, validate.In(AllApplicationRights())).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *CreateGatewayRequest) Validate() error {
	validations := make([]validate.Errors, 0, 4)
	validations = append(validations,
		validate.Field(req.Gateway.GatewayID, validate.ID).DescribeFieldName("Gateway ID"),
		validate.Field(req.Gateway.FrequencyPlanID, validate.Required).DescribeFieldName("Frequency plan ID"),
		validate.Field(req.Gateway.ClusterAddress, validate.Required).DescribeFieldName("Cluster Address"),
		validate.Field(req.OrganizationID, validate.NotRequired, validate.ID).DescribeFieldName("Organization ID"),
	)

	if req.Gateway.ContactAccount != nil {
		id := req.Gateway.ContactAccount.GetUserID()
		if id == "" {
			id = req.Gateway.ContactAccount.GetOrganizationID()
		}

		validations = append(validations, validate.Field(id, validate.NotRequired, validate.ID).DescribeFieldName("Contact account: Account ID"))
	}

	// if radios are set check for each one that frequency is present.
	for _, radio := range req.Gateway.Radios {
		validations = append(validations, validate.Field(radio.Frequency, validate.Required).DescribeFieldName("Radio Frequency"))
	}

	return validate.All(validations...)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateGatewayRequest) Validate() error {
	mask := req.GetUpdateMask()
	paths := mask.GetPaths()

	if paths == nil || len(paths) == 0 {
		return ErrEmptyUpdateMask.New(nil)
	}

	validations := make([]validate.Errors, 0)

	if err := validate.Field(req.Gateway.GatewayID, validate.ID).DescribeFieldName("Gateway ID"); err != nil {
		validations = append(validations, err)
	}

	var err validate.Errors
	for _, path := range paths {
		switch true {
		case FieldPathGatewayDescription.MatchString(path):
		case FieldPathGatewayPrivacySettingsStatusPublic.MatchString(path),
			FieldPathGatewayPrivacySettingsLocationPublic.MatchString(path),
			FieldPathGatewayPrivacySettingsContactable.MatchString(path),
			FieldPathGatewayAutoUpdate.MatchString(path),
			FieldPathGatewayPlatform.MatchString(path),
			FieldPathGatewayAntennas.MatchString(path),
			FieldPathGatewayAttributes.MatchString(path):
		case FieldPathGatewayClusterAddress.MatchString(path):
			err = validate.Field(req.Gateway.ClusterAddress, validate.Required).DescribeFieldName("Cluster Address")
		case FieldPathGatewayFrequencyPlanID.MatchString(path):
			err = validate.Field(req.Gateway.FrequencyPlanID, validate.Required).DescribeFieldName("Frequency plan ID")
		case FieldPathGatewayRadios.MatchString(path):
			for _, radio := range req.Gateway.Radios {
				validations = append(validations, validate.Field(radio.Frequency, validate.Required).DescribeFieldName("Radio Frequency"))
			}
		case FieldPathGatewayContactAccountUserID.MatchString(path):
			id := req.Gateway.ContactAccount.GetUserID()
			if id == "" {
				id = req.Gateway.ContactAccount.GetOrganizationID()
			}

			err = validate.Field(id, validate.NotRequired, validate.ID).DescribeFieldName("Contact account: Account ID")
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
func (req *ListGatewaysRequest) Validate() error {
	return validate.Field(req.OrganizationID, validate.NotRequired, validate.ID).DescribeFieldName("Organization ID")
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *GenerateGatewayAPIKeyRequest) Validate() error {
	return validate.All(
		validate.Field(req.GatewayID, validate.ID).DescribeFieldName("Gateway ID"),
		validate.Field(req.Name, validate.Required).DescribeFieldName("Key name"),
		validate.Field(req.Rights, validate.MinLength(1), validate.In(AllGatewayRights())).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateGatewayAPIKeyRequest) Validate() error {
	return validate.All(
		validate.Field(req.GatewayID, validate.ID).DescribeFieldName("Gateway ID"),
		validate.Field(req.Name, validate.Required).DescribeFieldName("Key name"),
		validate.Field(req.Rights, validate.MinLength(1), validate.In(AllGatewayRights())).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *RemoveGatewayAPIKeyRequest) Validate() error {
	return validate.All(
		validate.Field(req.GatewayID, validate.ID).DescribeFieldName("Gateway ID"),
		validate.Field(req.Name, validate.Required).DescribeFieldName("Key name"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *GatewayCollaborator) Validate() error {
	id := req.OrganizationOrUserIdentifier.GetUserID()
	if id == "" {
		id = req.OrganizationOrUserIdentifier.GetOrganizationID()
	}

	return validate.All(
		validate.Field(req.GatewayID, validate.ID).DescribeFieldName("Gateway ID"),
		validate.Field(id, validate.ID).DescribeFieldName("Account ID"),
		validate.Field(req.Rights, validate.NotRequired, validate.In(AllGatewayRights())).DescribeFieldName("Rights"),
	)
}

// validClientRights is the list of valid rights for a third-party client scope.
var validClientRights = []Right{
	RIGHT_USER_PROFILE_READ,
	RIGHT_USER_PROFILE_WRITE,
	RIGHT_USER_APPLICATIONS_LIST,
	RIGHT_USER_APPLICATIONS_CREATE,
	RIGHT_USER_GATEWAYS_LIST,
	RIGHT_USER_GATEWAYS_CREATE,
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
	paths := mask.GetPaths()

	if paths == nil || len(paths) == 0 {
		return ErrEmptyUpdateMask.New(nil)
	}

	validations := make([]validate.Errors, 0)

	if err := validate.Field(req.Client.ClientID, validate.ID).DescribeFieldName("Client ID"); err != nil {
		validations = append(validations, err)
	}

	var err validate.Errors
	for _, path := range paths {
		switch true {
		case FieldPathClientDescription.MatchString(path):
			err = validate.Field(req.Client.Description, validate.Required).DescribeFieldName("Description")
		case FieldPathClientRedirectURI.MatchString(path):
			err = validate.Field(req.Client.RedirectURI, validate.Required).DescribeFieldName("Redirect URI")
		case FieldPathClientRights.MatchString(path):
			err = validate.Field(req.Client.Rights, validate.MinLength(1), validate.In(validClientRights)).DescribeFieldName("Rights")
		default:
			return ErrInvalidPathUpdateMask.New(errors.Attributes{
				"path": path,
			})
		}

		validations = append(validations, err)
	}

	return validate.All(validations...)
}
