// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package storetest

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/pbkdf2"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// EntityAPIKey is an API key for an entity.
type EntityAPIKey struct {
	EntityIdentifiers *ttnpb.EntityIdentifiers
	APIKey            *ttnpb.APIKey
}

// EntityCollaborator is a collaborator on an entity.
type EntityCollaborator struct {
	EntityIdentifiers *ttnpb.EntityIdentifiers
	Collaborator      *ttnpb.Collaborator
}

// Population is a collection of store entities.
type Population struct {
	Applications  []*ttnpb.Application
	Clients       []*ttnpb.Client
	EndDevices    []*ttnpb.EndDevice
	Gateways      []*ttnpb.Gateway
	Organizations []*ttnpb.Organization
	Users         []*ttnpb.User
	UserSessions  []*ttnpb.UserSession

	APIKeys     []*EntityAPIKey
	Memberships []*EntityCollaborator
}

var now = time.Now()

// NewAPIKey adds a new API key to the population and returns it.
// The returned API key can not be modified and will not have its CreatedAt/UpdatedAt fields populated.
func (p *Population) NewAPIKey(entityID *ttnpb.EntityIdentifiers, rights ...ttnpb.Right) *ttnpb.APIKey {
	token, err := auth.APIKey.Generate(context.Background(), "")
	if err != nil {
		panic(err)
	}
	_, generatedID, generatedKey, err := auth.SplitToken(token)
	if err != nil {
		panic(err)
	}
	hashValidator := pbkdf2.Default()
	hashValidator.Iterations = 10
	hashedKey, err := auth.Hash(auth.NewContextWithHashValidator(context.Background(), hashValidator), generatedKey)
	if err != nil {
		panic(err)
	}
	p.APIKeys = append(p.APIKeys, &EntityAPIKey{
		EntityIdentifiers: entityID,
		APIKey: &ttnpb.APIKey{
			Id:     generatedID,
			Key:    hashedKey,
			Rights: rights,
		},
	})
	return &ttnpb.APIKey{
		Id:     generatedID,
		Key:    token,
		Rights: rights,
	}
}

// NewMembership adds a new membership to the population.
func (p *Population) NewMembership(memberID *ttnpb.OrganizationOrUserIdentifiers, entityID *ttnpb.EntityIdentifiers, rights ...ttnpb.Right) {
	if memberID == nil {
		return
	}
	p.Memberships = append(p.Memberships, &EntityCollaborator{
		EntityIdentifiers: entityID,
		Collaborator: &ttnpb.Collaborator{
			Ids:    memberID,
			Rights: rights,
		},
	})
}

// NewApplication adds a new application to the population and returns it.
// The returned application can be modified until Population.Populate is called.
func (p *Population) NewApplication(owner *ttnpb.OrganizationOrUserIdentifiers) *ttnpb.Application {
	i := len(p.Applications) + 1
	app := &ttnpb.Application{
		Ids: &ttnpb.ApplicationIdentifiers{
			ApplicationId: fmt.Sprintf("app-%02d", i),
		},
		Name: fmt.Sprintf("Application %02d", i),
	}
	p.Applications = append(p.Applications, app)
	p.NewMembership(owner, app.GetEntityIdentifiers(), ttnpb.RIGHT_ALL)
	return app
}

// NewClient adds a new client to the population and returns it.
// The returned client can be modified until Population.Populate is called.
func (p *Population) NewClient(owner *ttnpb.OrganizationOrUserIdentifiers) *ttnpb.Client {
	i := len(p.Clients) + 1
	cli := &ttnpb.Client{
		Ids: &ttnpb.ClientIdentifiers{
			ClientId: fmt.Sprintf("cli-%02d", i),
		},
		Name:  fmt.Sprintf("Client %02d", i),
		State: ttnpb.State_STATE_APPROVED,
	}
	p.Clients = append(p.Clients, cli)
	p.NewMembership(owner, cli.GetEntityIdentifiers(), ttnpb.RIGHT_ALL)
	return cli
}

// NewEndDevice adds a new end device to the population and returns it.
// The returned end device can be modified until Population.Populate is called.
func (p *Population) NewEndDevice(application *ttnpb.ApplicationIdentifiers) *ttnpb.EndDevice {
	i := len(p.EndDevices) + 1
	dev := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: application,
			DeviceId:       fmt.Sprintf("dev-%02d", i),
		},
		Name: fmt.Sprintf("End Device %02d", i),
	}
	p.EndDevices = append(p.EndDevices, dev)
	return dev
}

// NewGateway adds a new gateway to the population and returns it.
// The returned gateway can be modified until Population.Populate is called.
func (p *Population) NewGateway(owner *ttnpb.OrganizationOrUserIdentifiers) *ttnpb.Gateway {
	i := len(p.Gateways) + 1
	gtw := &ttnpb.Gateway{
		Ids: &ttnpb.GatewayIdentifiers{
			GatewayId: fmt.Sprintf("gtw-%02d", i),
		},
		Name: fmt.Sprintf("Gateway %02d", i),
	}
	p.Gateways = append(p.Gateways, gtw)
	p.NewMembership(owner, gtw.GetEntityIdentifiers(), ttnpb.RIGHT_ALL)
	return gtw
}

// NewOrganization adds a new organization to the population and returns it.
// The returned organization can be modified until Population.Populate is called.
func (p *Population) NewOrganization(owner *ttnpb.OrganizationOrUserIdentifiers) *ttnpb.Organization {
	i := len(p.Organizations) + 1
	org := &ttnpb.Organization{
		Ids: &ttnpb.OrganizationIdentifiers{
			OrganizationId: fmt.Sprintf("org-%02d", i),
		},
		Name: fmt.Sprintf("Organization %02d", i),
	}
	p.Organizations = append(p.Organizations, org)
	p.NewMembership(owner, org.GetEntityIdentifiers(), ttnpb.RIGHT_ALL)
	return org
}

// NewUser adds a new user to the population and returns it.
// The returned user can be modified until Population.Populate is called.
func (p *Population) NewUser() *ttnpb.User {
	i := len(p.Users) + 1
	usr := &ttnpb.User{
		Ids: &ttnpb.UserIdentifiers{
			UserId: fmt.Sprintf("usr-%02d", i),
		},
		Name:                           fmt.Sprintf("User %02d", i),
		PrimaryEmailAddress:            fmt.Sprintf("usr-%02d@example.com", i),
		PrimaryEmailAddressValidatedAt: ttnpb.ProtoTimePtr(now),
		State:                          ttnpb.State_STATE_APPROVED,
	}
	p.Users = append(p.Users, usr)
	return usr
}

// NewUserSession adds a new user session to the population and returns it.
// The returned user session can not be modified and will not have its CreatedAt/UpdatedAt fields populated.
func (p *Population) NewUserSession(user *ttnpb.UserIdentifiers) *ttnpb.UserSession {
	sessionID := uuid.New().String()
	generatedKey, err := auth.GenerateKey(context.Background())
	if err != nil {
		panic(err)
	}
	hashValidator := pbkdf2.Default()
	hashValidator.Iterations = 10
	hashedKey, err := auth.Hash(auth.NewContextWithHashValidator(context.Background(), hashValidator), generatedKey)
	if err != nil {
		panic(err)
	}
	p.UserSessions = append(p.UserSessions, &ttnpb.UserSession{
		UserIds:       user,
		SessionId:     sessionID,
		SessionSecret: hashedKey,
	})
	return &ttnpb.UserSession{
		UserIds:       user,
		SessionId:     sessionID,
		SessionSecret: generatedKey,
	}
}

// Populate creates the population in the database.
// After calling Populate, the entities in the population should no longer be modified.
func (p *Population) Populate(ctx context.Context, st Store) error {
	if len(p.Applications) > 0 {
		s, ok := st.(store.ApplicationStore)
		if !ok {
			return fmt.Errorf("store of type %T does not implement ApplicationStore", st)
		}
		for _, app := range p.Applications {
			created, err := s.CreateApplication(ctx, app)
			if err != nil {
				return err
			}
			*app = *created
		}
	}
	if len(p.Clients) > 0 {
		s, ok := st.(store.ClientStore)
		if !ok {
			return fmt.Errorf("store of type %T does not implement ClientStore", st)
		}
		for _, cli := range p.Clients {
			created, err := s.CreateClient(ctx, cli)
			if err != nil {
				return err
			}
			*cli = *created
		}
	}
	if len(p.EndDevices) > 0 {
		s, ok := st.(store.EndDeviceStore)
		if !ok {
			return fmt.Errorf("store of type %T does not implement EndDeviceStore", st)
		}
		for _, dev := range p.EndDevices {
			created, err := s.CreateEndDevice(ctx, dev)
			if err != nil {
				return err
			}
			*dev = *created
		}
	}
	if len(p.Gateways) > 0 {
		s, ok := st.(store.GatewayStore)
		if !ok {
			return fmt.Errorf("store of type %T does not implement GatewayStore", st)
		}
		for _, gtw := range p.Gateways {
			created, err := s.CreateGateway(ctx, gtw)
			if err != nil {
				return err
			}
			*gtw = *created
		}
	}
	if len(p.Organizations) > 0 {
		s, ok := st.(store.OrganizationStore)
		if !ok {
			return fmt.Errorf("store of type %T does not implement OrganizationStore", st)
		}
		for _, org := range p.Organizations {
			created, err := s.CreateOrganization(ctx, org)
			if err != nil {
				return err
			}
			*org = *created
		}
	}
	if len(p.Users) > 0 {
		s, ok := st.(store.UserStore)
		if !ok {
			return fmt.Errorf("store of type %T does not implement UserStore", st)
		}
		for _, usr := range p.Users {
			created, err := s.CreateUser(ctx, usr)
			if err != nil {
				return err
			}
			*usr = *created
		}
	}
	if len(p.UserSessions) > 0 {
		s, ok := st.(store.UserSessionStore)
		if !ok {
			return fmt.Errorf("store of type %T does not implement UserSessionStore", st)
		}
		for _, sess := range p.UserSessions {
			created, err := s.CreateSession(ctx, sess)
			if err != nil {
				return err
			}
			*sess = *created
		}
	}
	if len(p.APIKeys) > 0 {
		s, ok := st.(store.APIKeyStore)
		if !ok {
			return fmt.Errorf("store of type %T does not implement APIKeyStore", st)
		}
		for _, apiKey := range p.APIKeys {
			created, err := s.CreateAPIKey(ctx, apiKey.EntityIdentifiers, apiKey.APIKey)
			if err != nil {
				return err
			}
			*apiKey.APIKey = *created
		}
	}
	if len(p.Memberships) > 0 {
		s, ok := st.(store.MembershipStore)
		if !ok {
			return fmt.Errorf("store of type %T does not implement MembershipStore", st)
		}
		for _, collaborator := range p.Memberships {
			err := s.SetMember(ctx, collaborator.Collaborator.Ids, collaborator.EntityIdentifiers, &ttnpb.Rights{Rights: collaborator.Collaborator.GetRights()})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
