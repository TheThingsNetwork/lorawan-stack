// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestIsIDAllowed(t *testing.T) {
	a := assertions.New(t)

	settings := new(IdentityServerSettings)

	// all ids are allowed
	settings.BlacklistedIDs = nil
	a.So(settings.IsIDAllowed("foobar"), should.BeTrue)
	a.So(settings.IsIDAllowed("admin"), should.BeTrue)
	settings.BlacklistedIDs = []string{}
	a.So(settings.IsIDAllowed("foobar"), should.BeTrue)
	a.So(settings.IsIDAllowed("admin"), should.BeTrue)

	// `admin` is blacklisted
	settings.BlacklistedIDs = []string{"admin"}
	a.So(settings.IsIDAllowed("foobar"), should.BeTrue)
	a.So(settings.IsIDAllowed("admin"), should.BeFalse)
}

func TestSettingsValidations(t *testing.T) {
	a := assertions.New(t)

	{
		// empty request without update mask (bad)
		req := &UpdateSettingsRequest{}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrEmptyUpdateMask.Describes(err), should.BeTrue)

		// request with an invalid path in the update mask (bad)
		req = &UpdateSettingsRequest{
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"name", "foo"},
			},
		}
		err = req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrInvalidPathUpdateMask.Describes(err), should.BeTrue)

		// good request
		req = &UpdateSettingsRequest{
			Settings: IdentityServerSettings{
				IdentityServerSettings_UserRegistrationFlow: IdentityServerSettings_UserRegistrationFlow{
					SkipValidation: true,
				},
			},
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"user_registration.skip_validation"},
			},
		}
		a.So(req.Validate(), should.BeNil)
	}
}

func TestUserValidations(t *testing.T) {
	a := assertions.New(t)

	{
		// empty request (bad)
		req := &CreateUserRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// request with an invalid email (bad)
		req = &CreateUserRequest{
			User: User{
				UserIdentifier: UserIdentifier{UserID: "alice"},
				Name:           "Ali Ce",
				Password:       "12345678abC",
				Email:          "alice@alice.",
			},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &CreateUserRequest{
			User: User{
				UserIdentifier: UserIdentifier{UserID: "alice"},
				Name:           "Ali Ce",
				Password:       "12345678abC",
				Email:          "alice@alice.com",
			},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request without update mask (bad)
		req := &UpdateUserRequest{}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrEmptyUpdateMask.Describes(err), should.BeTrue)

		// request with an invalid path in the update mask (bad)
		req = &UpdateUserRequest{
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"name", "foo"},
			},
		}
		err = req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrInvalidPathUpdateMask.Describes(err), should.BeTrue)

		// good request
		req = &UpdateUserRequest{
			User: User{
				UserIdentifier: UserIdentifier{UserID: "alice"},
				Email:          "alice@ttn.com",
			},
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"name", "email"},
			},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &UpdateUserPasswordRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// password that dont have the minimum required length (bad)
		req = &UpdateUserPasswordRequest{
			Old: "lol",
			New: "1234",
		}
		a.So(req.Validate(), should.NotBeNil)

		// good password
		req = &UpdateUserPasswordRequest{
			Old: "lol",
			New: "fofofoof1234",
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &GenerateUserAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// empty list of rights (bad)
		req = &GenerateUserAPIKeyRequest{
			Name:   "foo",
			Rights: []Right{},
		}
		a.So(req.Validate(), should.NotBeNil)

		// request with gateway rights (bad)
		req = &GenerateUserAPIKeyRequest{
			Name:   "foo",
			Rights: []Right{RIGHT_GATEWAY_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &GenerateUserAPIKeyRequest{
			Name:   "foo",
			Rights: []Right{RIGHT_USER_APPLICATIONS_LIST},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &UpdateUserAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// request which tries to clear the rights (bad)
		req = &UpdateUserAPIKeyRequest{
			Name: "Foo-key",
		}
		a.So(req.Validate(), should.NotBeNil)

		// request with gateway rights (bad)
		req = &UpdateUserAPIKeyRequest{
			Name:   "Foo-key",
			Rights: []Right{RIGHT_GATEWAY_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &UpdateUserAPIKeyRequest{
			Name:   "Foo-key",
			Rights: []Right{RIGHT_USER_AUTHORIZEDCLIENTS},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &RemoveUserAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &RemoveUserAPIKeyRequest{
			Name: "foo",
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &ValidateUserEmailRequest{}
		a.So(req.Validate(), should.NotBeNil)

		req = &ValidateUserEmailRequest{
			Token: "foo",
		}
		a.So(req.Validate(), should.BeNil)
	}
}

func TestApplicationValidations(t *testing.T) {
	a := assertions.New(t)

	{
		// empty request (bad)
		req := &CreateApplicationRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &CreateApplicationRequest{
			Application: Application{
				ApplicationIdentifier: ApplicationIdentifier{ApplicationID: "foo-app"},
			},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// request without update mask (bad)
		req := &UpdateApplicationRequest{
			Application: Application{
				ApplicationIdentifier: ApplicationIdentifier{ApplicationID: "foo-app"},
			},
		}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrEmptyUpdateMask.Describes(err), should.BeTrue)

		// request with an invalid update mask (bad)
		req = &UpdateApplicationRequest{
			Application: Application{
				ApplicationIdentifier: ApplicationIdentifier{ApplicationID: "foo-app"},
			},
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"descriptio"},
			},
		}
		err = req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrInvalidPathUpdateMask.Describes(err), should.BeTrue)

		// good request
		req = &UpdateApplicationRequest{
			Application: Application{
				ApplicationIdentifier: ApplicationIdentifier{ApplicationID: "foo-app"},
			},
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"description"},
			},
		}
		err = req.Validate()
		a.So(err, should.BeNil)
	}

	{
		// empty request (bad)
		req := &GenerateApplicationAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// empty list of rights (bad)
		req = &GenerateApplicationAPIKeyRequest{
			ApplicationIdentifier: ApplicationIdentifier{ApplicationID: "foo-app"},
			Name:   "foo",
			Rights: []Right{},
		}
		a.So(req.Validate(), should.NotBeNil)

		// request with gateway rights (bad)
		req = &GenerateApplicationAPIKeyRequest{
			ApplicationIdentifier: ApplicationIdentifier{ApplicationID: "foo-app"},
			Name:   "foo",
			Rights: []Right{RIGHT_GATEWAY_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &GenerateApplicationAPIKeyRequest{
			ApplicationIdentifier: ApplicationIdentifier{ApplicationID: "foo-app"},
			Name:   "foo",
			Rights: []Right{RIGHT_APPLICATION_INFO},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &UpdateApplicationAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// request which tries to clear the rights (bad)
		req = &UpdateApplicationAPIKeyRequest{
			ApplicationIdentifier: ApplicationIdentifier{ApplicationID: "foo-app"},
			Name: "Foo-key",
		}
		a.So(req.Validate(), should.NotBeNil)

		// request with gateway rights (bad)
		req = &UpdateApplicationAPIKeyRequest{
			ApplicationIdentifier: ApplicationIdentifier{ApplicationID: "foo-app"},
			Name:   "foo",
			Rights: []Right{RIGHT_GATEWAY_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &UpdateApplicationAPIKeyRequest{
			ApplicationIdentifier: ApplicationIdentifier{ApplicationID: "foo-app"},
			Name:   "foo",
			Rights: []Right{RIGHT_APPLICATION_DELETE},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &RemoveApplicationAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &RemoveApplicationAPIKeyRequest{
			ApplicationIdentifier: ApplicationIdentifier{ApplicationID: "foo-app"},
			Name: "foo",
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &ApplicationCollaborator{}
		a.So(req.Validate(), should.NotBeNil)

		// request with gateway rights (bad)
		req = &ApplicationCollaborator{
			ApplicationIdentifier:        ApplicationIdentifier{"foo-app"},
			OrganizationOrUserIdentifier: OrganizationOrUserIdentifier{ID: &OrganizationOrUserIdentifier_UserID{"alice"}},
			Rights: []Right{RIGHT_GATEWAY_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &ApplicationCollaborator{
			ApplicationIdentifier:        ApplicationIdentifier{"foo-app"},
			OrganizationOrUserIdentifier: OrganizationOrUserIdentifier{ID: &OrganizationOrUserIdentifier_UserID{"alice"}},
		}
		a.So(req.Validate(), should.BeNil)
	}
}

func TestGatewayValidations(t *testing.T) {
	a := assertions.New(t)

	{
		// request with invalid gateway ID
		req := &CreateGatewayRequest{
			Gateway: Gateway{
				GatewayIdentifier: GatewayIdentifier{GatewayID: "__foo-gtw"},
				FrequencyPlanID:   "foo",
				ClusterAddress:    "foo",
			},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &CreateGatewayRequest{
			Gateway: Gateway{
				GatewayIdentifier: GatewayIdentifier{GatewayID: "foo-gtw"},
				FrequencyPlanID:   "foo",
				ClusterAddress:    "foo",
				Radios: []GatewayRadio{
					{
						Frequency: 12,
					},
				},
			},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// request without update mask (bad)
		req := &UpdateGatewayRequest{
			Gateway: Gateway{
				GatewayIdentifier: GatewayIdentifier{GatewayID: "__foo-gtw"},
				FrequencyPlanID:   "foo",
				ClusterAddress:    "foo",
			},
		}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrEmptyUpdateMask.Describes(err), should.BeTrue)

		// request with an invalid update mask (bad)
		req = &UpdateGatewayRequest{
			Gateway: Gateway{
				GatewayIdentifier: GatewayIdentifier{GatewayID: "__foo-gtw"},
				FrequencyPlanID:   "foo",
				ClusterAddress:    "foo",
			},
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"descriptio"},
			},
		}
		err = req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrInvalidPathUpdateMask.Describes(err), should.BeTrue)

		// good request
		req = &UpdateGatewayRequest{
			Gateway: Gateway{
				GatewayIdentifier: GatewayIdentifier{GatewayID: "foo-gtw"},
			},
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"description"},
			},
		}
		err = req.Validate()
		a.So(err, should.BeNil)
	}

	{
		// empty request (bad)
		req := &GenerateGatewayAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// empty list of rights (bad)
		req = &GenerateGatewayAPIKeyRequest{
			GatewayIdentifier: GatewayIdentifier{GatewayID: "foo-app"},
			Name:              "foo",
			Rights:            []Right{},
		}
		a.So(req.Validate(), should.NotBeNil)

		// rights for application (bad)
		req = &GenerateGatewayAPIKeyRequest{
			GatewayIdentifier: GatewayIdentifier{GatewayID: "foo-app"},
			Name:              "foo",
			Rights:            []Right{RIGHT_APPLICATION_INFO},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &GenerateGatewayAPIKeyRequest{
			GatewayIdentifier: GatewayIdentifier{GatewayID: "foo-app"},
			Name:              "foo",
			Rights:            []Right{RIGHT_GATEWAY_DELETE},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &UpdateGatewayAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// request which tries to clear the rights (bad)
		req = &UpdateGatewayAPIKeyRequest{
			GatewayIdentifier: GatewayIdentifier{GatewayID: "foo-app"},
			Name:              "Foo-key",
		}
		a.So(req.Validate(), should.NotBeNil)

		// request with application rights (bad)
		req = &UpdateGatewayAPIKeyRequest{
			GatewayIdentifier: GatewayIdentifier{GatewayID: "foo-app"},
			Name:              "foo",
			Rights:            []Right{RIGHT_APPLICATION_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &UpdateGatewayAPIKeyRequest{
			GatewayIdentifier: GatewayIdentifier{GatewayID: "foo-app"},
			Name:              "foo",
			Rights:            []Right{RIGHT_GATEWAY_DELETE},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &RemoveGatewayAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &RemoveGatewayAPIKeyRequest{
			GatewayIdentifier: GatewayIdentifier{GatewayID: "foo-app"},
			Name:              "foo",
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &GatewayCollaborator{}
		a.So(req.Validate(), should.NotBeNil)

		// request with application rights (bad)
		req = &GatewayCollaborator{
			GatewayIdentifier:            GatewayIdentifier{"foo-gtw"},
			OrganizationOrUserIdentifier: OrganizationOrUserIdentifier{ID: &OrganizationOrUserIdentifier_UserID{"alice"}},
			Rights: []Right{RIGHT_APPLICATION_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &GatewayCollaborator{
			GatewayIdentifier:            GatewayIdentifier{"foo-gtw"},
			OrganizationOrUserIdentifier: OrganizationOrUserIdentifier{ID: &OrganizationOrUserIdentifier_UserID{"alice"}},
		}
		a.So(req.Validate(), should.BeNil)
	}
}

func TestClientValidations(t *testing.T) {
	a := assertions.New(t)

	{
		// empty request (bad)
		req := &CreateClientRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &CreateClientRequest{
			Client: Client{
				Description:      "hi",
				ClientIdentifier: ClientIdentifier{ClientID: "foo-client"},
				RedirectURI:      "localhost",
				Rights:           []Right{RIGHT_APPLICATION_INFO},
			},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &UpdateClientRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// request without update_mask (bad)
		req = &UpdateClientRequest{
			Client: Client{
				ClientIdentifier: ClientIdentifier{ClientID: "foo-client"},
				Description:      "",
			},
		}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrEmptyUpdateMask.Describes(err), should.BeTrue)

		// request with invalid path fields on the update_mask (bad)
		req = &UpdateClientRequest{
			Client: Client{
				ClientIdentifier: ClientIdentifier{ClientID: "foo-client"},
				Description:      "foo description",
				RedirectURI:      "localhost",
				Rights:           []Right{RIGHT_APPLICATION_INFO},
			},
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"frequency_plan_id", "cluster_address"},
			},
		}
		err = req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrInvalidPathUpdateMask.Describes(err), should.BeTrue)

		// good request
		req = &UpdateClientRequest{
			Client: Client{
				ClientIdentifier: ClientIdentifier{ClientID: "foo-client"},
				Description:      "foo description",
				RedirectURI:      "ttn.com",
				Rights:           []Right{RIGHT_APPLICATION_INFO},
			},
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"redirect_uri", "rights", "description"},
			},
		}
		err = req.Validate()
		a.So(err, should.BeNil)
	}
}

func TestOrganizationValidations(t *testing.T) {
	a := assertions.New(t)

	{
		// request with invalid email
		req := &CreateOrganizationRequest{
			Organization: Organization{
				OrganizationIdentifier: OrganizationIdentifier{OrganizationID: "foo"},
				Email: "bar",
				Name:  "baz",
			},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &CreateOrganizationRequest{
			Organization: Organization{
				OrganizationIdentifier: OrganizationIdentifier{OrganizationID: "foo"},
				Email: "bar@bar.com",
				Name:  "baz",
			},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// request without update mask (bad)
		req := &UpdateOrganizationRequest{
			Organization: Organization{
				OrganizationIdentifier: OrganizationIdentifier{OrganizationID: "foo"},
				Name: "baz",
			},
		}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrEmptyUpdateMask.Describes(err), should.BeTrue)

		// request with an invalid update mask (bad)
		req = &UpdateOrganizationRequest{
			Organization: Organization{
				OrganizationIdentifier: OrganizationIdentifier{OrganizationID: "foo"},
				Name: "baz",
			},
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"descriptio"},
			},
		}
		err = req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrInvalidPathUpdateMask.Describes(err), should.BeTrue)

		// request with good update mask but invalid email
		req = &UpdateOrganizationRequest{
			Organization: Organization{
				OrganizationIdentifier: OrganizationIdentifier{OrganizationID: "foo"},
			},
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"email"},
			},
		}
		err = req.Validate()
		a.So(err, should.NotBeNil)

		// good request
		req = &UpdateOrganizationRequest{
			Organization: Organization{
				OrganizationIdentifier: OrganizationIdentifier{OrganizationID: "foo"},
			},
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"description"},
			},
		}
		err = req.Validate()
		a.So(err, should.BeNil)
	}

	{
		// empty request (bad)
		req := &GenerateOrganizationAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// empty list of rights (bad)
		req = &GenerateOrganizationAPIKeyRequest{
			OrganizationIdentifier: OrganizationIdentifier{OrganizationID: "foo"},
			Name:   "foo",
			Rights: []Right{},
		}
		a.So(req.Validate(), should.NotBeNil)

		// rights for application (bad)
		req = &GenerateOrganizationAPIKeyRequest{
			OrganizationIdentifier: OrganizationIdentifier{OrganizationID: "foo"},
			Name:   "foo",
			Rights: []Right{RIGHT_APPLICATION_INFO},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &GenerateOrganizationAPIKeyRequest{
			OrganizationIdentifier: OrganizationIdentifier{OrganizationID: "foo"},
			Name:   "foo",
			Rights: []Right{RIGHT_ORGANIZATION_DELETE},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &UpdateOrganizationAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// request which tries to clear the rights (bad)
		req = &UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifier: OrganizationIdentifier{OrganizationID: "foo"},
			Name: "Foo-key",
		}
		a.So(req.Validate(), should.NotBeNil)

		// request with application rights (bad)
		req = &UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifier: OrganizationIdentifier{OrganizationID: "foo"},
			Name:   "foo",
			Rights: []Right{RIGHT_APPLICATION_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifier: OrganizationIdentifier{OrganizationID: "foo"},
			Name:   "foo",
			Rights: []Right{RIGHT_ORGANIZATION_INFO},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &RemoveOrganizationAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &RemoveOrganizationAPIKeyRequest{
			OrganizationIdentifier: OrganizationIdentifier{OrganizationID: "foo"},
			Name: "foo",
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &OrganizationMember{}
		a.So(req.Validate(), should.NotBeNil)

		// request with application rights (bad)
		req = &OrganizationMember{
			OrganizationIdentifier: OrganizationIdentifier{OrganizationID: "foo"},
			UserIdentifier:         UserIdentifier{UserID: "alice"},
			Rights:                 []Right{RIGHT_APPLICATION_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &OrganizationMember{
			OrganizationIdentifier: OrganizationIdentifier{OrganizationID: "foo"},
			UserIdentifier:         UserIdentifier{UserID: "alice"},
		}
		a.So(req.Validate(), should.BeNil)
	}
}
