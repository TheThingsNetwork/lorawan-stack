// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package store

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/auth/pbkdf2"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/randutil"
)

// NewPopulator returns a new database populator with a population of the given size.
// It is seeded by the given seed.
func NewPopulator(size int, seed int64) *Populator {
	randy := rand.New(randutil.NewLockedSource(rand.NewSource(seed)))
	p := &Populator{
		APIKeys:     make(map[*ttnpb.EntityIdentifiers][]*ttnpb.APIKey),
		Memberships: make(map[*ttnpb.EntityIdentifiers][]*ttnpb.Collaborator),
	}
	for i := 0; i < size; i++ {
		application := ttnpb.NewPopulatedApplication(randy, false)
		application.Description = fmt.Sprintf("Random Application %d", i+1)
		applicationID := application.EntityIdentifiers()
		p.Applications = append(p.Applications, application)
		p.APIKeys[applicationID] = append(
			p.APIKeys[applicationID],
			&ttnpb.APIKey{
				Name:   "default key",
				Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
			},
		)
		client := ttnpb.NewPopulatedClient(randy, false)
		client.Description = fmt.Sprintf("Random Client %d", i+1)
		p.Clients = append(p.Clients, client)
		gateway := ttnpb.NewPopulatedGateway(randy, false)
		gateway.Description = fmt.Sprintf("Random Gateway %d", i+1)
		gatewayID := gateway.EntityIdentifiers()
		p.Gateways = append(p.Gateways, gateway)
		p.APIKeys[gatewayID] = append(
			p.APIKeys[gatewayID],
			&ttnpb.APIKey{
				Name:   "default key",
				Rights: []ttnpb.Right{ttnpb.RIGHT_GATEWAY_ALL},
			},
		)
		organization := ttnpb.NewPopulatedOrganization(randy, false)
		organization.Description = fmt.Sprintf("Random Organization %d", i+1)
		organizationID := organization.EntityIdentifiers()
		p.Organizations = append(p.Organizations, organization)
		p.APIKeys[organizationID] = append(
			p.APIKeys[organizationID],
			&ttnpb.APIKey{
				Name:   "default key",
				Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL, ttnpb.RIGHT_CLIENT_ALL, ttnpb.RIGHT_GATEWAY_ALL, ttnpb.RIGHT_ORGANIZATION_ALL},
			},
		)
		user := ttnpb.NewPopulatedUser(randy, false)
		user.Description = fmt.Sprintf("Random User %d", i+1)
		userID := user.EntityIdentifiers()
		p.Users = append(p.Users, user)
		p.APIKeys[userID] = append(
			p.APIKeys[userID],
			&ttnpb.APIKey{
				Name:   "default key",
				Rights: []ttnpb.Right{ttnpb.RIGHT_ALL},
			},
		)
	}
	var userIndex, organizationIndex int
	for _, application := range p.Applications {
		applicationID := application.EntityIdentifiers()
		userCollaborators := randy.Intn((len(p.Users)/10)+1) + 1
		for i := 0; i < userCollaborators; i++ {
			ouID := p.Users[userIndex%len(p.Users)].OrganizationOrUserIdentifiers()
			p.Memberships[applicationID] = append(p.Memberships[applicationID], &ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *ouID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
			})
			userIndex++
		}
		organizationCollaborators := randy.Intn((len(p.Organizations)/10)+1) + 1
		for i := 0; i < organizationCollaborators; i++ {
			ouID := p.Organizations[organizationIndex%len(p.Organizations)].OrganizationOrUserIdentifiers()
			p.Memberships[applicationID] = append(p.Memberships[applicationID], &ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *ouID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
			})
			organizationIndex++
		}
	}
	for _, client := range p.Clients {
		clientID := client.EntityIdentifiers()
		userCollaborators := randy.Intn((len(p.Users)/10)+1) + 1
		for i := 0; i < userCollaborators; i++ {
			ouID := p.Users[userIndex%len(p.Users)].OrganizationOrUserIdentifiers()
			p.Memberships[clientID] = append(p.Memberships[clientID], &ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *ouID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_CLIENT_ALL},
			})
			userIndex++
		}
		organizationCollaborators := randy.Intn((len(p.Organizations)/10)+1) + 1
		for i := 0; i < organizationCollaborators; i++ {
			ouID := p.Organizations[organizationIndex%len(p.Organizations)].OrganizationOrUserIdentifiers()
			p.Memberships[clientID] = append(p.Memberships[clientID], &ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *ouID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_CLIENT_ALL},
			})
			organizationIndex++
		}
	}
	for _, gateway := range p.Gateways {
		gatewayID := gateway.EntityIdentifiers()
		userCollaborators := randy.Intn((len(p.Users)/10)+1) + 1
		for i := 0; i < userCollaborators; i++ {
			ouID := p.Users[userIndex%len(p.Users)].OrganizationOrUserIdentifiers()
			p.Memberships[gatewayID] = append(p.Memberships[gatewayID], &ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *ouID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_GATEWAY_ALL},
			})
			userIndex++
		}
		organizationCollaborators := randy.Intn((len(p.Organizations)/10)+1) + 1
		for i := 0; i < organizationCollaborators; i++ {
			ouID := p.Organizations[organizationIndex%len(p.Organizations)].OrganizationOrUserIdentifiers()
			p.Memberships[gatewayID] = append(p.Memberships[gatewayID], &ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *ouID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_GATEWAY_ALL},
			})
			organizationIndex++
		}
	}
	for _, organization := range p.Organizations {
		organizationID := organization.EntityIdentifiers()
		userCollaborators := randy.Intn((len(p.Users)/10)+1) + 1
		for i := 0; i < userCollaborators; i++ {
			ouID := p.Users[userIndex%len(p.Users)].OrganizationOrUserIdentifiers()
			p.Memberships[organizationID] = append(p.Memberships[organizationID], &ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *ouID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL, ttnpb.RIGHT_CLIENT_ALL, ttnpb.RIGHT_GATEWAY_ALL, ttnpb.RIGHT_ORGANIZATION_ALL},
			})
			userIndex++
		}
	}
	return p
}

// Populator is intended to populate a database with test data.
type Populator struct {
	Applications  []*ttnpb.Application
	Clients       []*ttnpb.Client
	Gateways      []*ttnpb.Gateway
	Organizations []*ttnpb.Organization
	Users         []*ttnpb.User

	APIKeys     map[*ttnpb.EntityIdentifiers][]*ttnpb.APIKey
	Memberships map[*ttnpb.EntityIdentifiers][]*ttnpb.Collaborator
}

// Populate the database.
func (p *Populator) Populate(ctx context.Context, db *gorm.DB) (err error) {
	hashValidator := pbkdf2.Default()
	hashValidator.Iterations = 10
	ctx = auth.NewContextWithHashValidator(ctx, hashValidator)
	if err = p.populateApplications(ctx, db); err != nil {
		return err
	}
	if err = p.populateClients(ctx, db); err != nil {
		return err
	}
	if err = p.populateGateways(ctx, db); err != nil {
		return err
	}
	if err = p.populateOrganizations(ctx, db); err != nil {
		return err
	}
	if err = p.populateUsers(ctx, db); err != nil {
		return err
	}
	if err = p.populateAPIKeys(ctx, db); err != nil {
		return err
	}
	if err = p.populateMemberships(ctx, db); err != nil {
		return err
	}
	return nil
}

func (p *Populator) populateApplications(ctx context.Context, db *gorm.DB) (err error) {
	for i, application := range p.Applications {
		p.Applications[i], err = GetApplicationStore(db).CreateApplication(ctx, application)
		if err != nil {
			return err
		}
		p.Applications[i].ContactInfo, err = GetContactInfoStore(db).SetContactInfo(ctx, application.EntityIdentifiers(), application.ContactInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Populator) populateClients(ctx context.Context, db *gorm.DB) (err error) {
	for i, client := range p.Clients {
		secret := client.Secret
		hashedSecret, _ := auth.Hash(ctx, client.Secret)
		client.Secret = hashedSecret
		p.Clients[i], err = GetClientStore(db).CreateClient(ctx, client)
		if err != nil {
			return err
		}
		client.Secret = secret
		p.Clients[i].ContactInfo, err = GetContactInfoStore(db).SetContactInfo(ctx, client.EntityIdentifiers(), client.ContactInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Populator) populateGateways(ctx context.Context, db *gorm.DB) (err error) {
	for i, gateway := range p.Gateways {
		p.Gateways[i], err = GetGatewayStore(db).CreateGateway(ctx, gateway)
		if err != nil {
			return err
		}
		p.Gateways[i].ContactInfo, err = GetContactInfoStore(db).SetContactInfo(ctx, gateway.EntityIdentifiers(), gateway.ContactInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Populator) populateOrganizations(ctx context.Context, db *gorm.DB) (err error) {
	for i, organization := range p.Organizations {
		p.Organizations[i], err = GetOrganizationStore(db).CreateOrganization(ctx, organization)
		if err != nil {
			return err
		}
		p.Organizations[i].ContactInfo, err = GetContactInfoStore(db).SetContactInfo(ctx, organization.EntityIdentifiers(), organization.ContactInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Populator) populateUsers(ctx context.Context, db *gorm.DB) (err error) {
	for i, user := range p.Users {
		password := user.Password
		hashedPassword, _ := auth.Hash(ctx, user.Password)
		user.Password = hashedPassword
		p.Users[i], err = GetUserStore(db).CreateUser(ctx, user)
		if err != nil {
			return err
		}
		p.Users[i].Password = password
		p.Users[i].ContactInfo, err = GetContactInfoStore(db).SetContactInfo(ctx, user.EntityIdentifiers(), user.ContactInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Populator) populateAPIKeys(ctx context.Context, db *gorm.DB) (err error) {
	for entityID, apiKeys := range p.APIKeys {
		for _, apiKey := range apiKeys {
			token, _ := auth.APIKey.Generate(ctx, "")
			_, generatedID, generatedKey, _ := auth.SplitToken(token)
			hashedKey, _ := auth.Hash(ctx, generatedKey)
			apiKey.ID = generatedID
			apiKey.Key = hashedKey
			if err = GetAPIKeyStore(db).CreateAPIKey(ctx, entityID, apiKey); err != nil {
				return err
			}
			apiKey.Key = token
		}
	}
	return nil
}

func (p *Populator) populateMemberships(ctx context.Context, db *gorm.DB) (err error) {
	for entityID, members := range p.Memberships {
		for _, member := range members {
			if err = GetMembershipStore(db).SetMember(
				ctx,
				&member.OrganizationOrUserIdentifiers,
				entityID,
				ttnpb.RightsFrom(member.Rights...),
			); err != nil {
				return err
			}
		}
	}
	return nil
}
