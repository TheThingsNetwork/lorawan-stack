// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"testing"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestUserValidations(t *testing.T) {
	a := assertions.New(t)

	{
		req := &CreateUserRequest{}
		a.So(req.Validate(), should.NotBeNil)

		req = &CreateUserRequest{
			UserIdentifier: UserIdentifier{"alice"},
			Name:           "Ali Ce",
			Password:       "12345678abC",
			Email:          "alice@alice.",
		}
		a.So(req.Validate(), should.NotBeNil)

		req.Email = "alice@alice.com"
		a.So(req.Validate(), should.BeNil)
	}

	{
		req := &UpdateUserRequest{}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrUpdateMaskNotFound.Describes(err), should.BeTrue)

		req = &UpdateUserRequest{
			UpdateMask: &ptypes.FieldMask{
				Paths: []string{"name", "foo"},
			},
		}
		err = req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrInvalidPathUpdateMask.Describes(err), should.BeTrue)

		req = &UpdateUserRequest{
			UpdateMask: &ptypes.FieldMask{
				Paths: []string{"name", "email"},
			},
			Email: "alice@ttn.com",
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		req := &UpdateUserPasswordRequest{}
		a.So(req.Validate(), should.NotBeNil)

		req.New = "1234"
		a.So(req.Validate(), should.NotBeNil)

		req.New = "abcABC123494949"
		a.So(req.Validate(), should.BeNil)
	}
}

func TestApplicationValidations(t *testing.T) {
	a := assertions.New(t)

	{
		req := &CreateApplicationRequest{}
		a.So(req.Validate(), should.NotBeNil)

		req = &CreateApplicationRequest{
			ApplicationIdentifier: ApplicationIdentifier{"foo-app"},
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		req := &UpdateApplicationRequest{}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrUpdateMaskNotFound.Describes(err), should.BeTrue)

		req = &UpdateApplicationRequest{
			UpdateMask: &ptypes.FieldMask{
				Paths: []string{"descriptio"},
			},
		}
		err = req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrInvalidPathUpdateMask.Describes(err), should.BeTrue)

		req = &UpdateApplicationRequest{
			ApplicationIdentifier: ApplicationIdentifier{"foo-app"},
			UpdateMask: &ptypes.FieldMask{
				Paths: []string{"description"},
			},
		}
		err = req.Validate()
		a.So(err, should.BeNil)
	}

	{
		req := &GenerateApplicationAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		req = &GenerateApplicationAPIKeyRequest{
			KeyName: "foo",
			Rights:  []Right{},
		}
		a.So(req.Validate(), should.NotBeNil)

		req = &GenerateApplicationAPIKeyRequest{
			KeyName: "foo",
			Rights:  []Right{Right(1)},
		}
		a.So(req.Validate(), should.BeNil)

	}

	{
		req := &RemoveApplicationAPIKeyRequest{}
		a.So(req.Validate(), should.NotBeNil)

		req = &RemoveApplicationAPIKeyRequest{
			KeyName: "foo",
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		req := &SetApplicationCollaboratorRequest{}
		a.So(req.Validate(), should.NotBeNil)

		req = &SetApplicationCollaboratorRequest{
			ApplicationIdentifier: ApplicationIdentifier{"foo-app"},
			Collaborator: Collaborator{
				UserIdentifier: UserIdentifier{"alice"},
			},
		}
	}
}

func TestGatewayValidations(t *testing.T) {
	a := assertions.New(t)

	{
		req := &CreateGatewayRequest{
			GatewayIdentifier: GatewayIdentifier{"__foo-gtw"},
			FrequencyPlanID:   "foo",
			ClusterAddress:    "foo",
		}
		a.So(req.Validate(), should.NotBeNil)

		req = &CreateGatewayRequest{
			GatewayIdentifier: GatewayIdentifier{"foo-gtw"},
			FrequencyPlanID:   "foo",
			ClusterAddress:    "foo",
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		req := &UpdateGatewayRequest{}
		a.So(req.Validate(), should.NotBeNil)

		req = &UpdateGatewayRequest{
			FrequencyPlanID: "",
			ClusterAddress:  "localhost:1234",
			UpdateMask: &ptypes.FieldMask{
				Paths: []string{"frequency_plan_id", "cluster_address"},
			},
		}
		a.So(req.Validate(), should.NotBeNil)

		req = &UpdateGatewayRequest{
			FrequencyPlanID: "fooPlan",
			ClusterAddress:  "localhost:1234",
			UpdateMask: &ptypes.FieldMask{
				Paths: []string{"frequency_plan_id", "cluster_addressss"},
			},
		}
		err := req.Validate()
		a.So(err, should.NotBeNil)
		a.So(ErrInvalidPathUpdateMask.Describes(err), should.BeTrue)

		req = &UpdateGatewayRequest{
			GatewayIdentifier: GatewayIdentifier{"foo-gtw"},
			FrequencyPlanID:   "fooPlan",
			ClusterAddress:    "localhost:1234",
			UpdateMask: &ptypes.FieldMask{
				Paths: []string{"frequency_plan_id", "cluster_address"},
			},
		}
		err = req.Validate()
		a.So(err, should.BeNil)
	}

	{
		req := &SetGatewayCollaboratorRequest{}
		a.So(req.Validate(), should.NotBeNil)

		req = &SetGatewayCollaboratorRequest{
			GatewayIdentifier: GatewayIdentifier{"foo-gtw"},
			Collaborator: Collaborator{
				UserIdentifier: UserIdentifier{"alice"},
			},
		}
	}
}

func TestClientValidations(t *testing.T) {
	a := assertions.New(t)

	{
		req := &SetClientStateRequest{}
		a.So(req.Validate(), should.NotBeNil)

		req = &SetClientStateRequest{
			ClientIdentifier: ClientIdentifier{"foo-client"},
			State:            ClientState(5),
		}
		a.So(req.Validate(), should.NotBeNil)

		req = &SetClientStateRequest{
			ClientIdentifier: ClientIdentifier{"foo-client"},
			State:            ClientState(0),
		}
		a.So(req.Validate(), should.BeNil)
	}

	{
		req := &SetClientCollaboratorRequest{}
		a.So(req.Validate(), should.NotBeNil)

		req = &SetClientCollaboratorRequest{
			ClientIdentifier: ClientIdentifier{"foo-client"},
			Collaborator: Collaborator{
				UserIdentifier: UserIdentifier{"alice"},
			},
		}
	}

}
