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
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestSettingsValidations(t *testing.T) {
	a := assertions.New(t)

	{
		// empty request without update mask (bad)
		req := &UpdateSettingsRequest{}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(err, should.HaveSameErrorDefinitionAs, errMissingUpdateMask)

		// request with an invalid path in the update mask (bad)
		req = &UpdateSettingsRequest{
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"name", "foo"},
			},
		}
		err = req.Validate()
		a.So(err, should.NotBeNil)
		a.So(err, should.HaveSameErrorDefinitionAs, errInvalidPathUpdateMask)

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
				UserIdentifiers: UserIdentifiers{
					UserID: "alice",
					Email:  "alice@alice.",
				},
				Name:     "Ali Ce",
				Password: "12345678abC",
			},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &CreateUserRequest{
			User: User{
				UserIdentifiers: UserIdentifiers{
					UserID: "alice",
					Email:  "alice@alice.com",
				},
				Name:     "Ali Ce",
				Password: "12345678abC",
			},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request without update mask (bad)
		req := &UpdateUserRequest{}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(err, should.HaveSameErrorDefinitionAs, errMissingUpdateMask)

		// request with an invalid path in the update mask (bad)
		req = &UpdateUserRequest{
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"name", "foo"},
			},
		}
		err = req.Validate()
		a.So(err, should.NotBeNil)
		a.So(err, should.HaveSameErrorDefinitionAs, errInvalidPathUpdateMask)

		// good request
		req = &UpdateUserRequest{
			User: User{
				UserIdentifiers: UserIdentifiers{
					UserID: "alice",
					Email:  "alice@ttn.com",
				},
			},
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"name", "ids.email"},
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
			Rights: []Right{RIGHT_USER_AUTHORIZED_CLIENTS},
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
				ApplicationIdentifiers: ApplicationIdentifiers{ApplicationID: "foo-app"},
			},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// request without update mask (bad)
		req := &UpdateApplicationRequest{
			Application: Application{
				ApplicationIdentifiers: ApplicationIdentifiers{ApplicationID: "foo-app"},
			},
		}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(err, should.HaveSameErrorDefinitionAs, errMissingUpdateMask)

		// request with an invalid update mask (bad)
		req = &UpdateApplicationRequest{
			Application: Application{
				ApplicationIdentifiers: ApplicationIdentifiers{ApplicationID: "foo-app"},
			},
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"descriptio"},
			},
		}
		err = req.Validate()
		a.So(err, should.NotBeNil)
		a.So(err, should.HaveSameErrorDefinitionAs, errInvalidPathUpdateMask)

		// good request
		req = &UpdateApplicationRequest{
			Application: Application{
				ApplicationIdentifiers: ApplicationIdentifiers{ApplicationID: "foo-app"},
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
			ApplicationIdentifiers: ApplicationIdentifiers{ApplicationID: "foo-app"},
			Name:   "foo",
			Rights: []Right{},
		}
		a.So(req.Validate(), should.NotBeNil)

		// request with gateway rights (bad)
		req = &GenerateApplicationAPIKeyRequest{
			ApplicationIdentifiers: ApplicationIdentifiers{ApplicationID: "foo-app"},
			Name:   "foo",
			Rights: []Right{RIGHT_GATEWAY_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &GenerateApplicationAPIKeyRequest{
			ApplicationIdentifiers: ApplicationIdentifiers{ApplicationID: "foo-app"},
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
			ApplicationIdentifiers: ApplicationIdentifiers{ApplicationID: "foo-app"},
			Name: "Foo-key",
		}
		a.So(req.Validate(), should.NotBeNil)

		// request with gateway rights (bad)
		req = &UpdateApplicationAPIKeyRequest{
			ApplicationIdentifiers: ApplicationIdentifiers{ApplicationID: "foo-app"},
			Name:   "foo",
			Rights: []Right{RIGHT_GATEWAY_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &UpdateApplicationAPIKeyRequest{
			ApplicationIdentifiers: ApplicationIdentifiers{ApplicationID: "foo-app"},
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
			ApplicationIdentifiers: ApplicationIdentifiers{ApplicationID: "foo-app"},
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
			ApplicationIdentifiers:        ApplicationIdentifiers{"foo-app"},
			OrganizationOrUserIdentifiers: OrganizationOrUserIdentifiers{ID: &OrganizationOrUserIdentifiers_UserID{UserID: &UserIdentifiers{UserID: "alice"}}},
			Rights: []Right{RIGHT_GATEWAY_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &ApplicationCollaborator{
			ApplicationIdentifiers:        ApplicationIdentifiers{"foo-app"},
			OrganizationOrUserIdentifiers: OrganizationOrUserIdentifiers{ID: &OrganizationOrUserIdentifiers_UserID{UserID: &UserIdentifiers{UserID: "alice"}}},
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
				GatewayIdentifiers: GatewayIdentifiers{GatewayID: "__foo-gtw"},
				FrequencyPlanID:    "foo",
				ClusterAddress:     "foo",
			},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &CreateGatewayRequest{
			Gateway: Gateway{
				GatewayIdentifiers: GatewayIdentifiers{GatewayID: "foo-gtw"},
				FrequencyPlanID:    "foo",
				ClusterAddress:     "foo",
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
				GatewayIdentifiers: GatewayIdentifiers{GatewayID: "__foo-gtw"},
				FrequencyPlanID:    "foo",
				ClusterAddress:     "foo",
			},
		}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(err, should.HaveSameErrorDefinitionAs, errMissingUpdateMask)

		// request with an invalid update mask (bad)
		req = &UpdateGatewayRequest{
			Gateway: Gateway{
				GatewayIdentifiers: GatewayIdentifiers{GatewayID: "__foo-gtw"},
				FrequencyPlanID:    "foo",
				ClusterAddress:     "foo",
			},
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"descriptio"},
			},
		}
		err = req.Validate()
		a.So(err, should.NotBeNil)
		a.So(err, should.HaveSameErrorDefinitionAs, errInvalidPathUpdateMask)

		// good request
		req = &UpdateGatewayRequest{
			Gateway: Gateway{
				GatewayIdentifiers: GatewayIdentifiers{GatewayID: "foo-gtw"},
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
			GatewayIdentifiers: GatewayIdentifiers{GatewayID: "foo-app"},
			Name:               "foo",
			Rights:             []Right{},
		}
		a.So(req.Validate(), should.NotBeNil)

		// rights for application (bad)
		req = &GenerateGatewayAPIKeyRequest{
			GatewayIdentifiers: GatewayIdentifiers{GatewayID: "foo-app"},
			Name:               "foo",
			Rights:             []Right{RIGHT_APPLICATION_INFO},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &GenerateGatewayAPIKeyRequest{
			GatewayIdentifiers: GatewayIdentifiers{GatewayID: "foo-app"},
			Name:               "foo",
			Rights:             []Right{RIGHT_GATEWAY_DELETE},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &UpdateGatewayAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// request which tries to clear the rights (bad)
		req = &UpdateGatewayAPIKeyRequest{
			GatewayIdentifiers: GatewayIdentifiers{GatewayID: "foo-app"},
			Name:               "Foo-key",
		}
		a.So(req.Validate(), should.NotBeNil)

		// request with application rights (bad)
		req = &UpdateGatewayAPIKeyRequest{
			GatewayIdentifiers: GatewayIdentifiers{GatewayID: "foo-app"},
			Name:               "foo",
			Rights:             []Right{RIGHT_APPLICATION_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &UpdateGatewayAPIKeyRequest{
			GatewayIdentifiers: GatewayIdentifiers{GatewayID: "foo-app"},
			Name:               "foo",
			Rights:             []Right{RIGHT_GATEWAY_DELETE},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &RemoveGatewayAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &RemoveGatewayAPIKeyRequest{
			GatewayIdentifiers: GatewayIdentifiers{GatewayID: "foo-app"},
			Name:               "foo",
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &GatewayCollaborator{}
		a.So(req.Validate(), should.NotBeNil)

		// request with application rights (bad)
		req = &GatewayCollaborator{
			GatewayIdentifiers:            GatewayIdentifiers{GatewayID: "foo-gtw"},
			OrganizationOrUserIdentifiers: OrganizationOrUserIdentifiers{ID: &OrganizationOrUserIdentifiers_UserID{UserID: &UserIdentifiers{UserID: "alice"}}},
			Rights: []Right{RIGHT_APPLICATION_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &GatewayCollaborator{
			GatewayIdentifiers:            GatewayIdentifiers{GatewayID: "foo-gtw"},
			OrganizationOrUserIdentifiers: OrganizationOrUserIdentifiers{ID: &OrganizationOrUserIdentifiers_UserID{UserID: &UserIdentifiers{UserID: "alice"}}},
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
				Description:       "hi",
				ClientIdentifiers: ClientIdentifiers{ClientID: "foo-client"},
				RedirectURI:       "localhost",
				Rights:            []Right{RIGHT_APPLICATION_INFO},
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
				ClientIdentifiers: ClientIdentifiers{ClientID: "foo-client"},
				Description:       "",
			},
		}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(err, should.HaveSameErrorDefinitionAs, errMissingUpdateMask)

		// request with invalid path fields on the update_mask (bad)
		req = &UpdateClientRequest{
			Client: Client{
				ClientIdentifiers: ClientIdentifiers{ClientID: "foo-client"},
				Description:       "foo description",
				RedirectURI:       "localhost",
				Rights:            []Right{RIGHT_APPLICATION_INFO},
			},
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"frequency_plan_id", "cluster_address"},
			},
		}
		err = req.Validate()
		a.So(err, should.NotBeNil)
		a.So(err, should.HaveSameErrorDefinitionAs, errInvalidPathUpdateMask)

		// good request
		req = &UpdateClientRequest{
			Client: Client{
				ClientIdentifiers: ClientIdentifiers{ClientID: "foo-client"},
				Description:       "foo description",
				RedirectURI:       "ttn.com",
				Rights:            []Right{RIGHT_APPLICATION_INFO},
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
				OrganizationIdentifiers: OrganizationIdentifiers{OrganizationID: "foo"},
				Email: "bar",
				Name:  "baz",
			},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &CreateOrganizationRequest{
			Organization: Organization{
				OrganizationIdentifiers: OrganizationIdentifiers{OrganizationID: "foo"},
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
				OrganizationIdentifiers: OrganizationIdentifiers{OrganizationID: "foo"},
				Name: "baz",
			},
		}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(err, should.HaveSameErrorDefinitionAs, errMissingUpdateMask)

		// request with an invalid update mask (bad)
		req = &UpdateOrganizationRequest{
			Organization: Organization{
				OrganizationIdentifiers: OrganizationIdentifiers{OrganizationID: "foo"},
				Name: "baz",
			},
			UpdateMask: pbtypes.FieldMask{
				Paths: []string{"descriptio"},
			},
		}
		err = req.Validate()
		a.So(err, should.NotBeNil)
		a.So(err, should.HaveSameErrorDefinitionAs, errInvalidPathUpdateMask)

		// request with good update mask but invalid email
		req = &UpdateOrganizationRequest{
			Organization: Organization{
				OrganizationIdentifiers: OrganizationIdentifiers{OrganizationID: "foo"},
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
				OrganizationIdentifiers: OrganizationIdentifiers{OrganizationID: "foo"},
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
			OrganizationIdentifiers: OrganizationIdentifiers{OrganizationID: "foo"},
			Name:   "foo",
			Rights: []Right{},
		}
		a.So(req.Validate(), should.NotBeNil)

		// rights for application (bad)
		req = &GenerateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: OrganizationIdentifiers{OrganizationID: "foo"},
			Name:   "foo",
			Rights: []Right{RIGHT_APPLICATION_INFO},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &GenerateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: OrganizationIdentifiers{OrganizationID: "foo"},
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
			OrganizationIdentifiers: OrganizationIdentifiers{OrganizationID: "foo"},
			Name: "Foo-key",
		}
		a.So(req.Validate(), should.NotBeNil)

		// request with application rights (bad)
		req = &UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: OrganizationIdentifiers{OrganizationID: "foo"},
			Name:   "foo",
			Rights: []Right{RIGHT_APPLICATION_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: OrganizationIdentifiers{OrganizationID: "foo"},
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
			OrganizationIdentifiers: OrganizationIdentifiers{OrganizationID: "foo"},
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
			OrganizationIdentifiers: OrganizationIdentifiers{OrganizationID: "foo"},
			UserIdentifiers:         UserIdentifiers{UserID: "alice"},
			Rights:                  []Right{RIGHT_APPLICATION_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &OrganizationMember{
			OrganizationIdentifiers: OrganizationIdentifiers{OrganizationID: "foo"},
			UserIdentifiers:         UserIdentifiers{UserID: "alice"},
		}
		a.So(req.Validate(), should.BeNil)
	}
}
