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

import (
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/validate"
)

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateSettingsRequest) Validate() error {
	mask := req.GetUpdateMask()
	paths := mask.GetPaths()

	if len(paths) == 0 {
		return ErrEmptyUpdateMask.New(nil)
	}

	validations := make([]error, 0)

	var err error
	for _, path := range paths {
		switch true {
		case FieldPathSettingsBlacklistedIDs.MatchString(path):
			for _, id := range req.Settings.BlacklistedIDs {
				err = validate.Field(id, validate.ID).DescribeFieldName("Blacklisted ID")
				if err != nil {
					validations = append(validations, err)
				}
			}
		case FieldPathSettingsUserRegistrationSkipValidation.MatchString(path),
			FieldPathSettingsUserRegistrationInvitationOnly.MatchString(path),
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
		validate.Field(req.User.UserIdentifiers.UserID, validate.ID).DescribeFieldName("User ID"),
		validate.Field(req.User.UserIdentifiers.Email, validate.Email).DescribeFieldName("Email"),
		validate.Field(req.User.Password, validate.Password).DescribeFieldName("Password"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateUserRequest) Validate() error {
	mask := req.GetUpdateMask()
	paths := mask.GetPaths()

	if len(paths) == 0 {
		return ErrEmptyUpdateMask.New(nil)
	}

	validations := make([]error, 0)

	var err error
	for _, path := range paths {
		switch true {
		case FieldPathUserName.MatchString(path),
			FieldPathUserAdmin.MatchString(path),
			FieldPathUserRequirePasswordUpdate.MatchString(path):
		case FieldPathUserEmail.MatchString(path):
			err = validate.Field(req.User.UserIdentifiers.Email, validate.Email).DescribeFieldName("Email")
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

	if len(paths) == 0 {
		return ErrEmptyUpdateMask.New(nil)
	}

	validations := make([]error, 0)

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

// List of valid rights for an application collaborator.
// `RIGHT_APPLICATION_LINK` is excluded.
var validApplicationCollaboratorRights = DifferenceRights(AllApplicationRights(), []Right{RIGHT_APPLICATION_LINK})

// Validate is used as validator function by the GRPC validator interceptor.
func (req *ApplicationCollaborator) Validate() error {
	return validate.All(
		req.OrganizationOrUserIdentifiers.Validate(),
		req.ApplicationIdentifiers.Validate(),
		validate.Field(req.Rights, validate.NotRequired, validate.In(validApplicationCollaboratorRights)).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *CreateGatewayRequest) Validate() error {
	validations := make([]error, 0, 4)
	validations = append(validations,
		validate.Field(req.Gateway.GatewayIdentifiers.GatewayID, validate.ID).DescribeFieldName("Gateway ID"),
		validate.Field(req.Gateway.GatewayIdentifiers.EUI, validate.NotRequired).DescribeFieldName("EUI"),
		validate.Field(req.Gateway.FrequencyPlanID, validate.Required).DescribeFieldName("Frequency plan ID"),
		validate.Field(req.Gateway.ClusterAddress, validate.Required).DescribeFieldName("Cluster Address"),
		validate.Field(req.OrganizationID, validate.NotRequired, validate.ID).DescribeFieldName("Organization ID"),
	)

	if req.Gateway.ContactAccountIDs != nil {
		validations = append(validations, req.Gateway.ContactAccountIDs.Validate())
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

	if len(paths) == 0 {
		return ErrEmptyUpdateMask.New(nil)
	}

	validations := make([]error, 0)

	if err := req.Gateway.GatewayIdentifiers.Validate(); err != nil {
		validations = append(validations, err)
	}

	var err error
	for _, path := range paths {
		switch true {
		case FieldPathGatewayDescription.MatchString(path):
		case FieldPathGatewayPrivacySettingsStatusPublic.MatchString(path),
			FieldPathGatewayPrivacySettingsLocationPublic.MatchString(path),
			FieldPathGatewayPrivacySettingsContactable.MatchString(path),
			FieldPathGatewayAutoUpdate.MatchString(path),
			FieldPathGatewayPlatform.MatchString(path),
			FieldPathGatewayAntennas.MatchString(path),
			FieldPathGatewayAttributes.MatchString(path),
			FieldPathGatewayDisableTxDelay.MatchString(path):
		case FieldPathGatewayClusterAddress.MatchString(path):
			err = validate.Field(req.Gateway.ClusterAddress, validate.Required).DescribeFieldName("Cluster Address")
		case FieldPathGatewayFrequencyPlanID.MatchString(path):
			err = validate.Field(req.Gateway.FrequencyPlanID, validate.Required).DescribeFieldName("Frequency plan ID")
		case FieldPathGatewayRadios.MatchString(path):
			for _, radio := range req.Gateway.Radios {
				validations = append(validations, validate.Field(radio.Frequency, validate.Required).DescribeFieldName("Radio Frequency"))
			}
		case FieldPathGatewayContactAccountIDs.MatchString(path):
			if req.Gateway.ContactAccountIDs == nil {
				continue
			}

			err = req.Gateway.ContactAccountIDs.Validate()
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
		req.GatewayIdentifiers.Validate(),
		validate.Field(req.Name, validate.Required).DescribeFieldName("Key name"),
		validate.Field(req.Rights, validate.MinLength(1), validate.In(AllGatewayRights())).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateGatewayAPIKeyRequest) Validate() error {
	return validate.All(
		req.GatewayIdentifiers.Validate(),
		validate.Field(req.Name, validate.Required).DescribeFieldName("Key name"),
		validate.Field(req.Rights, validate.MinLength(1), validate.In(AllGatewayRights())).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *RemoveGatewayAPIKeyRequest) Validate() error {
	return validate.All(
		req.GatewayIdentifiers.Validate(),
		validate.Field(req.Name, validate.Required).DescribeFieldName("Key name"),
	)
}

// List of valid rights for a gateway collaborator.
// `RIGHT_GATEWAY_LINK` is excluded.
var validGatewayCollaboratorRights = DifferenceRights(AllGatewayRights(), []Right{RIGHT_GATEWAY_LINK})

// Validate is used as validator function by the GRPC validator interceptor.
func (req *GatewayCollaborator) Validate() error {
	return validate.All(
		req.OrganizationOrUserIdentifiers.Validate(),
		req.GatewayIdentifiers.Validate(),
		validate.Field(req.Rights, validate.NotRequired, validate.In(validGatewayCollaboratorRights)).DescribeFieldName("Rights"),
	)
}

// validClientRights is the list of valid rights for a third-party client scope.
// Rights related to API keys, third party clients and self-user account deletion
// are excluded.
var validClientRights = []Right{
	RIGHT_USER_INFO,
	RIGHT_USER_SETTINGS_BASIC,
	RIGHT_USER_APPLICATIONS_LIST,
	RIGHT_USER_APPLICATIONS_CREATE,
	RIGHT_USER_GATEWAYS_LIST,
	RIGHT_USER_GATEWAYS_CREATE,
	RIGHT_USER_ORGANIZATIONS_LIST,
	RIGHT_USER_ORGANIZATIONS_CREATE,
	RIGHT_APPLICATION_INFO,
	RIGHT_APPLICATION_SETTINGS_BASIC,
	RIGHT_APPLICATION_SETTINGS_COLLABORATORS,
	RIGHT_APPLICATION_DELETE,
	RIGHT_APPLICATION_DEVICES_READ,
	RIGHT_APPLICATION_DEVICES_WRITE,
	RIGHT_APPLICATION_TRAFFIC_READ,
	RIGHT_APPLICATION_TRAFFIC_UP_WRITE,
	RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE,
	RIGHT_GATEWAY_INFO,
	RIGHT_GATEWAY_SETTINGS_BASIC,
	RIGHT_GATEWAY_SETTINGS_COLLABORATORS,
	RIGHT_GATEWAY_DELETE,
	RIGHT_GATEWAY_TRAFFIC_READ,
	RIGHT_GATEWAY_TRAFFIC_DOWN_WRITE,
	RIGHT_GATEWAY_STATUS_READ,
	RIGHT_GATEWAY_LOCATION_READ,
	RIGHT_ORGANIZATION_INFO,
	RIGHT_ORGANIZATION_SETTINGS_BASIC,
	RIGHT_ORGANIZATION_SETTINGS_MEMBERS,
	RIGHT_ORGANIZATION_DELETE,
	RIGHT_ORGANIZATION_APPLICATIONS_CREATE,
	RIGHT_ORGANIZATION_APPLICATIONS_LIST,
	RIGHT_ORGANIZATION_GATEWAYS_CREATE,
	RIGHT_ORGANIZATION_GATEWAYS_LIST,
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *CreateClientRequest) Validate() error {
	return validate.All(
		req.Client.ClientIdentifiers.Validate(),
		validate.Field(req.Client.Description, validate.Required).DescribeFieldName("Description"),
		validate.Field(req.Client.RedirectURI, validate.Required).DescribeFieldName("Redirect URI"),
		validate.Field(req.Client.Rights, validate.MinLength(1), validate.In(validClientRights)).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateClientRequest) Validate() error {
	mask := req.GetUpdateMask()
	paths := mask.GetPaths()

	if len(paths) == 0 {
		return ErrEmptyUpdateMask.New(nil)
	}

	validations := make([]error, 0)

	if err := validate.Field(req.Client.ClientID, validate.ID).DescribeFieldName("Client ID"); err != nil {
		validations = append(validations, err)
	}

	var err error
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

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateOrganizationRequest) Validate() error {
	mask := req.GetUpdateMask()
	paths := mask.GetPaths()

	if len(paths) == 0 {
		return ErrEmptyUpdateMask.New(nil)
	}

	validations := make([]error, 0)
	validations = append(validations, validate.Field(req.Organization.OrganizationID, validate.ID).DescribeFieldName("Organization ID"))

	var err error
	for _, path := range paths {
		switch true {
		case FieldPathOrganizationName.MatchString(path),
			FieldPathOrganizationDescription.MatchString(path),
			FieldPathOrganizationURL.MatchString(path),
			FieldPathOrganizationLocation.MatchString(path):
		case FieldPathOrganizationEmail.MatchString(path):
			err = validate.Field(req.Organization.Email, validate.Email).DescribeFieldName("Email")
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
func (req *CreateOrganizationRequest) Validate() error {
	return validate.All(
		validate.Field(req.Organization.OrganizationID, validate.ID).DescribeFieldName("Organization ID"),
		validate.Field(req.Organization.Email, validate.Email).DescribeFieldName("Email"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *GenerateOrganizationAPIKeyRequest) Validate() error {
	return validate.All(
		validate.Field(req.OrganizationID, validate.ID).DescribeFieldName("Organization ID"),
		validate.Field(req.Name, validate.Required).DescribeFieldName("Key name"),
		validate.Field(req.Rights, validate.MinLength(1), validate.In(AllOrganizationRights())).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *UpdateOrganizationAPIKeyRequest) Validate() error {
	return validate.All(
		validate.Field(req.OrganizationID, validate.ID).DescribeFieldName("Organization ID"),
		validate.Field(req.Name, validate.Required).DescribeFieldName("Key name"),
		validate.Field(req.Rights, validate.MinLength(1), validate.In(AllOrganizationRights())).DescribeFieldName("Rights"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *RemoveOrganizationAPIKeyRequest) Validate() error {
	return validate.All(
		validate.Field(req.OrganizationID, validate.ID).DescribeFieldName("Organization ID"),
		validate.Field(req.Name, validate.Required).DescribeFieldName("Key name"),
	)
}

// Validate is used as validator function by the GRPC validator interceptor.
func (req *OrganizationMember) Validate() error {
	return validate.All(
		validate.Field(req.OrganizationID, validate.ID).DescribeFieldName("Organization ID"),
		validate.Field(req.UserID, validate.ID).DescribeFieldName("User ID"),
		validate.Field(req.Rights, validate.NotRequired, validate.In(AllOrganizationRights())).DescribeFieldName("Rights"),
	)
}
