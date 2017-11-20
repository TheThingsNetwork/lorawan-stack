// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestSettingsMask(t *testing.T) {
	a := assertions.New(t)

	mask := new(IdentityServerSettingsMask)
	mask.FullMask()
	a.So(mask, should.Resemble, &IdentityServerSettingsMask{
		BlacklistedIDs:     true,
		AutomaticApproval:  true,
		ClosedRegistration: true,
		ValidationTokenTTL: true,
		AllowedEmails:      true,
	})
}

func TestSettingsValidations(t *testing.T) {
	a := assertions.New(t)

	// empty update mask (bad)
	req := &UpdateSettingsRequest{}
	a.So(ErrEmptyUpdateMask.Describes(req.Validate()), should.BeTrue)

	// empty update mask (bad)
	req = &UpdateSettingsRequest{
		UpdateMask: IdentityServerSettingsMask{},
	}
	a.So(ErrEmptyUpdateMask.Describes(req.Validate()), should.BeTrue)

	// request with not valid ids (bad)
	req = &UpdateSettingsRequest{
		Settings: IdentityServerSettings{
			BlacklistedIDs: []string{"s", "webui"},
		},
		UpdateMask: IdentityServerSettingsMask{
			BlacklistedIDs: true,
		},
	}
	a.So(req.Validate(), should.NotBeNil)

	// good request
	req = &UpdateSettingsRequest{
		Settings: IdentityServerSettings{
			BlacklistedIDs: []string{"webui", "self"},
		},
		UpdateMask: IdentityServerSettingsMask{
			BlacklistedIDs: true,
		},
	}
	a.So(req.Validate(), should.BeNil)

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
				UserIdentifier: UserIdentifier{"alice"},
				Name:           "Ali Ce",
				Password:       "12345678abC",
				Email:          "alice@alice.",
			},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &CreateUserRequest{
			User: User{
				UserIdentifier: UserIdentifier{"alice"},
				Name:           "Ali Ce",
				Password:       "12345678abC",
				Email:          "alice@alice.com",
			},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request with empty update mask (bad)
		req := &UpdateUserRequest{}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrEmptyUpdateMask.Describes(err), should.BeTrue)

		// good request
		req = &UpdateUserRequest{
			User: User{
				Email: "alice@ttn.com",
			},
			UpdateMask: UserMask{
				Name:  true,
				Email: true,
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
				ApplicationIdentifier: ApplicationIdentifier{"foo-app"},
			},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// request with empty update mask (bad)
		req := &UpdateApplicationRequest{
			Application: Application{
				ApplicationIdentifier: ApplicationIdentifier{"foo-app"},
			},
		}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrEmptyUpdateMask.Describes(err), should.BeTrue)

		// good request
		req = &UpdateApplicationRequest{
			Application: Application{
				ApplicationIdentifier: ApplicationIdentifier{"foo-app"},
			},
			UpdateMask: ApplicationMask{
				Description: true,
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
			ApplicationIdentifier: ApplicationIdentifier{"foo-app"},
			Name:   "foo",
			Rights: []Right{},
		}
		a.So(req.Validate(), should.NotBeNil)

		// request with gateway rights (bad)
		req = &GenerateApplicationAPIKeyRequest{
			ApplicationIdentifier: ApplicationIdentifier{"foo-app"},
			Name:   "foo",
			Rights: []Right{RIGHT_GATEWAY_DELETE},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &GenerateApplicationAPIKeyRequest{
			ApplicationIdentifier: ApplicationIdentifier{"foo-app"},
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
			ApplicationIdentifier: ApplicationIdentifier{"foo-app"},
			Key: APIKey{
				Name: "Foo-key",
			},
			UpdateMask: APIKeyMask{
				Name:   true,
				Rights: true,
			},
		}
		a.So(req.Validate(), should.NotBeNil)

		// request with empty update mask (bad)
		req = &UpdateApplicationAPIKeyRequest{
			ApplicationIdentifier: ApplicationIdentifier{"foo-app"},
			Key: APIKey{
				Name: "Foo-key",
			},
		}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrEmptyUpdateMask.Describes(err), should.BeTrue)

		// request with gateway rights (bad)
		req = &UpdateApplicationAPIKeyRequest{
			ApplicationIdentifier: ApplicationIdentifier{"foo-app"},
			Key: APIKey{
				Name:   "Foo-key",
				Rights: []Right{RIGHT_GATEWAY_DELETE},
			},
			UpdateMask: APIKeyMask{
				Name:   true,
				Rights: true,
			},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &UpdateApplicationAPIKeyRequest{
			ApplicationIdentifier: ApplicationIdentifier{"foo-app"},
			Key: APIKey{
				Name: "Foo-key",
			},
			UpdateMask: APIKeyMask{
				Name: true,
			},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &RemoveApplicationAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &RemoveApplicationAPIKeyRequest{
			ApplicationIdentifier: ApplicationIdentifier{"foo-app"},
			Key: "foo",
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &SetApplicationCollaboratorRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// request with gateway rights (bad)
		req = &SetApplicationCollaboratorRequest{
			ApplicationIdentifier: ApplicationIdentifier{"foo-app"},
			Collaborator: Collaborator{
				UserIdentifier: UserIdentifier{"alice"},
				Rights:         []Right{RIGHT_GATEWAY_DELETE},
			},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &SetApplicationCollaboratorRequest{
			ApplicationIdentifier: ApplicationIdentifier{"foo-app"},
			Collaborator: Collaborator{
				UserIdentifier: UserIdentifier{"alice"},
			},
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
				GatewayIdentifier: GatewayIdentifier{"__foo-gtw"},
				FrequencyPlanID:   "foo",
				ClusterAddress:    "foo",
			},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &CreateGatewayRequest{
			Gateway: Gateway{
				GatewayIdentifier: GatewayIdentifier{"foo-gtw"},
				FrequencyPlanID:   "foo",
				ClusterAddress:    "foo",
			},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request without update_mask
		req := &UpdateGatewayRequest{}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrEmptyUpdateMask.Describes(err), should.BeTrue)

		// request with empty frequency plan ID and no gateway ID (bad)
		req = &UpdateGatewayRequest{
			Gateway: Gateway{
				FrequencyPlanID: "",
				ClusterAddress:  "foo",
			},
			UpdateMask: GatewayMask{
				ClusterAddress:  true,
				FrequencyPlanID: true,
			},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &UpdateGatewayRequest{
			Gateway: Gateway{
				GatewayIdentifier: GatewayIdentifier{"foo-gtw"},
				FrequencyPlanID:   "foo",
				ClusterAddress:    "foo",
			},
			UpdateMask: GatewayMask{
				ClusterAddress:  true,
				FrequencyPlanID: true,
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
			GatewayIdentifier: GatewayIdentifier{"foo-app"},
			Name:              "foo",
			Rights:            []Right{},
		}
		a.So(req.Validate(), should.NotBeNil)

		// rights for application (bad)
		req = &GenerateGatewayAPIKeyRequest{
			GatewayIdentifier: GatewayIdentifier{"foo-app"},
			Name:              "foo",
			Rights:            []Right{RIGHT_APPLICATION_INFO},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &GenerateGatewayAPIKeyRequest{
			GatewayIdentifier: GatewayIdentifier{"foo-app"},
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
			GatewayIdentifier: GatewayIdentifier{"foo-app"},
			Key: APIKey{
				Name: "Foo-key",
			},
			UpdateMask: APIKeyMask{
				Name:   true,
				Rights: true,
			}}
		a.So(req.Validate(), should.NotBeNil)

		// request with empty update mask (bad)
		req = &UpdateGatewayAPIKeyRequest{
			GatewayIdentifier: GatewayIdentifier{"foo-app"},
			Key: APIKey{
				Name: "Foo-key",
			},
		}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrEmptyUpdateMask.Describes(err), should.BeTrue)

		// request with application rights (bad)
		req = &UpdateGatewayAPIKeyRequest{
			GatewayIdentifier: GatewayIdentifier{"foo-app"},
			Key: APIKey{
				Name:   "Foo-key",
				Rights: []Right{RIGHT_APPLICATION_DELETE},
			},
			UpdateMask: APIKeyMask{
				Name:   true,
				Rights: true,
			},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &UpdateGatewayAPIKeyRequest{
			GatewayIdentifier: GatewayIdentifier{"foo-app"},
			Key: APIKey{
				Name: "Foo-key",
			},
			UpdateMask: APIKeyMask{
				Name: true,
			},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &RemoveGatewayAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &RemoveGatewayAPIKeyRequest{
			GatewayIdentifier: GatewayIdentifier{"foo-app"},
			Key:               "foo",
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		// empty request (bad)
		req := &SetGatewayCollaboratorRequest{}
		a.So(req.Validate(), should.NotBeNil)

		// request with application rights (bad)
		req = &SetGatewayCollaboratorRequest{
			GatewayIdentifier: GatewayIdentifier{"foo-gtw"},
			Collaborator: Collaborator{
				UserIdentifier: UserIdentifier{"alice"},
				Rights:         []Right{RIGHT_APPLICATION_DELETE},
			},
		}
		a.So(req.Validate(), should.NotBeNil)

		// good request
		req = &SetGatewayCollaboratorRequest{
			GatewayIdentifier: GatewayIdentifier{"foo-gtw"},
			Collaborator: Collaborator{
				UserIdentifier: UserIdentifier{"alice"},
			},
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
				ClientIdentifier: ClientIdentifier{"foo-client"},
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

		// request with empty update_mask (bad)
		req = &UpdateClientRequest{
			Client: Client{
				ClientIdentifier: ClientIdentifier{"foo-client"},
				Description:      "",
			},
		}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrEmptyUpdateMask.Describes(err), should.BeTrue)

		// good request
		req = &UpdateClientRequest{
			Client: Client{
				ClientIdentifier: ClientIdentifier{"foo-client"},
				Description:      "ho",
				RedirectURI:      "ttn.com",
				Rights:           []Right{RIGHT_APPLICATION_INFO},
			},
			UpdateMask: ClientMask{
				RedirectURI: true,
				Description: true,
				Rights:      true,
			},
		}
		err = req.Validate()
		a.So(err, should.BeNil)
	}
}
