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
	"time"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/pbkdf2"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// NewPopulator returns a new database populator with a population of the given size.
// It is seeded by the given seed.
func NewPopulator(size int, seed int64) *Populator {
	now := time.Now()
	randy := rand.New(rand.NewSource(seed))
	p := &Populator{
		APIKeys:     make(map[*ttnpb.EntityIdentifiers][]*ttnpb.APIKey),
		Memberships: make(map[*ttnpb.EntityIdentifiers][]*ttnpb.Collaborator),
	}
	for i := 0; i < size; i++ {
		application := &ttnpb.Application{
			Ids:         &ttnpb.ApplicationIdentifiers{ApplicationId: fmt.Sprintf("random-app-%d", i+1)},
			Name:        fmt.Sprintf("Random %d", i+1),
			Description: fmt.Sprintf("Randomly generated Application %d", i+1),
		}
		applicationID := application.GetEntityIdentifiers()
		p.Applications = append(p.Applications, application)
		p.APIKeys[applicationID] = append(
			p.APIKeys[applicationID],
			&ttnpb.APIKey{
				Name:   "default key",
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_ALL},
			},
		)
		client := &ttnpb.Client{
			Ids:         &ttnpb.ClientIdentifiers{ClientId: fmt.Sprintf("random-cli-%d", i+1)},
			Name:        fmt.Sprintf("Random %d", i+1),
			Description: fmt.Sprintf("Randomly generated Client %d", i+1),
		}
		p.Clients = append(p.Clients, client)
		gateway := &ttnpb.Gateway{
			Ids:         &ttnpb.GatewayIdentifiers{GatewayId: fmt.Sprintf("random-gtw-%d", i+1)},
			Name:        fmt.Sprintf("Random %d", i+1),
			Description: fmt.Sprintf("Randomly generated Gateway %d", i+1),
		}
		gatewayID := gateway.GetEntityIdentifiers()

		// This is to prevent the IS trying to use the randomly generated key IDs to decrypt the secrets.
		if gateway.LbsLnsSecret != nil {
			gateway.LbsLnsSecret.KeyId = ""
		}
		if gateway.ClaimAuthenticationCode != nil && gateway.ClaimAuthenticationCode.Secret != nil {
			gateway.ClaimAuthenticationCode.Secret.KeyId = ""
		}
		if gateway.TargetCupsKey != nil {
			gateway.TargetCupsKey.KeyId = ""
		}

		p.Gateways = append(p.Gateways, gateway)
		p.APIKeys[gatewayID] = append(
			p.APIKeys[gatewayID],
			&ttnpb.APIKey{
				Name:   "default key",
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_ALL},
			},
		)
		organization := &ttnpb.Organization{
			Ids:         &ttnpb.OrganizationIdentifiers{OrganizationId: fmt.Sprintf("random-org-%d", i+1)},
			Name:        fmt.Sprintf("Random %d", i+1),
			Description: fmt.Sprintf("Randomly generated Organization %d", i+1),
		}
		organizationID := organization.GetEntityIdentifiers()
		p.Organizations = append(p.Organizations, organization)
		p.APIKeys[organizationID] = append(
			p.APIKeys[organizationID],
			&ttnpb.APIKey{
				Name:   "default key",
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_ALL, ttnpb.Right_RIGHT_CLIENT_ALL, ttnpb.Right_RIGHT_GATEWAY_ALL, ttnpb.Right_RIGHT_ORGANIZATION_ALL},
			},
		)
		user := &ttnpb.User{
			Ids:                            &ttnpb.UserIdentifiers{UserId: fmt.Sprintf("random-usr-%d", i+1)},
			Name:                           fmt.Sprintf("Random %d", i+1),
			Description:                    fmt.Sprintf("Randomly generated User %d", i+1),
			PrimaryEmailAddress:            fmt.Sprintf("user-%d@example.com", i+1),
			PrimaryEmailAddressValidatedAt: ttnpb.ProtoTimePtr(now),
		}
		userID := user.GetEntityIdentifiers()
		p.Users = append(p.Users, user)
		p.APIKeys[userID] = append(
			p.APIKeys[userID],
			&ttnpb.APIKey{
				Name:   "default key",
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_ALL},
			},
		)
		eui := &types.EUI64{}
		eui.UnmarshalNumber(uint64(i + 1))
		endDevice := &ttnpb.EndDevice{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: application.Ids,
				JoinEui:        eui,
				DevEui:         eui,
				DeviceId:       fmt.Sprintf("random-device-%d", i+1),
			},
		}
		p.EndDevices = append(p.EndDevices, endDevice)
	}
	var userIndex, organizationIndex int
	for _, application := range p.Applications {
		applicationID := application.GetEntityIdentifiers()
		userCollaborators := randy.Intn((len(p.Users)/10)+1) + 1
		for i := 0; i < userCollaborators; i++ {
			ouID := p.Users[userIndex%len(p.Users)].OrganizationOrUserIdentifiers()
			p.Memberships[applicationID] = append(p.Memberships[applicationID], &ttnpb.Collaborator{
				Ids:    ouID,
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_ALL},
			})
			userIndex++
		}
		organizationCollaborators := randy.Intn((len(p.Organizations)/10)+1) + 1
		for i := 0; i < organizationCollaborators; i++ {
			ouID := p.Organizations[organizationIndex%len(p.Organizations)].OrganizationOrUserIdentifiers()
			p.Memberships[applicationID] = append(p.Memberships[applicationID], &ttnpb.Collaborator{
				Ids:    ouID,
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_ALL},
			})
			organizationIndex++
		}
	}
	for _, client := range p.Clients {
		clientID := client.GetEntityIdentifiers()
		userCollaborators := randy.Intn((len(p.Users)/10)+1) + 1
		for i := 0; i < userCollaborators; i++ {
			ouID := p.Users[userIndex%len(p.Users)].OrganizationOrUserIdentifiers()
			p.Memberships[clientID] = append(p.Memberships[clientID], &ttnpb.Collaborator{
				Ids:    ouID,
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_CLIENT_ALL},
			})
			userIndex++
		}
		organizationCollaborators := randy.Intn((len(p.Organizations)/10)+1) + 1
		for i := 0; i < organizationCollaborators; i++ {
			ouID := p.Organizations[organizationIndex%len(p.Organizations)].OrganizationOrUserIdentifiers()
			p.Memberships[clientID] = append(p.Memberships[clientID], &ttnpb.Collaborator{
				Ids:    ouID,
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_CLIENT_ALL},
			})
			organizationIndex++
		}
	}
	for _, gateway := range p.Gateways {
		gatewayID := gateway.GetEntityIdentifiers()
		userCollaborators := randy.Intn((len(p.Users)/10)+1) + 1
		for i := 0; i < userCollaborators; i++ {
			ouID := p.Users[userIndex%len(p.Users)].OrganizationOrUserIdentifiers()
			p.Memberships[gatewayID] = append(p.Memberships[gatewayID], &ttnpb.Collaborator{
				Ids:    ouID,
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_ALL},
			})
			userIndex++
		}
		organizationCollaborators := randy.Intn((len(p.Organizations)/10)+1) + 1
		for i := 0; i < organizationCollaborators; i++ {
			ouID := p.Organizations[organizationIndex%len(p.Organizations)].OrganizationOrUserIdentifiers()
			p.Memberships[gatewayID] = append(p.Memberships[gatewayID], &ttnpb.Collaborator{
				Ids:    ouID,
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_ALL},
			})
			organizationIndex++
		}
	}
	for _, organization := range p.Organizations {
		organizationID := organization.GetEntityIdentifiers()
		userCollaborators := randy.Intn((len(p.Users)/10)+1) + 1
		for i := 0; i < userCollaborators; i++ {
			ouID := p.Users[userIndex%len(p.Users)].OrganizationOrUserIdentifiers()
			p.Memberships[organizationID] = append(p.Memberships[organizationID], &ttnpb.Collaborator{
				Ids:    ouID,
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_ALL, ttnpb.Right_RIGHT_CLIENT_ALL, ttnpb.Right_RIGHT_GATEWAY_ALL, ttnpb.Right_RIGHT_ORGANIZATION_ALL},
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
	EndDevices    []*ttnpb.EndDevice

	APIKeys     map[*ttnpb.EntityIdentifiers][]*ttnpb.APIKey
	Memberships map[*ttnpb.EntityIdentifiers][]*ttnpb.Collaborator
}

// Populate the database.
func (p *Populator) Populate(ctx context.Context, db *gorm.DB) (err error) {
	tx := db.Begin()
	defer func() {
		if commitErr := tx.Commit().Error; err == nil {
			err = commitErr
		}
	}()

	hashValidator := pbkdf2.Default()
	hashValidator.Iterations = 10
	ctx = auth.NewContextWithHashValidator(ctx, hashValidator)
	if err = p.populateApplications(ctx, tx); err != nil {
		return fmt.Errorf("failed to populate applications: %w", err)
	}
	if err = p.populateClients(ctx, tx); err != nil {
		return fmt.Errorf("failed to populate clients: %w", err)
	}
	if err = p.populateGateways(ctx, tx); err != nil {
		return fmt.Errorf("failed to populate gateways: %w", err)
	}
	if err = p.populateOrganizations(ctx, tx); err != nil {
		return fmt.Errorf("failed to populate organizations: %w", err)
	}
	if err = p.populateUsers(ctx, tx); err != nil {
		return fmt.Errorf("failed to populate users: %w", err)
	}
	if err = p.populateAPIKeys(ctx, tx); err != nil {
		return fmt.Errorf("failed to populate API keys: %w", err)
	}
	if err = p.populateMemberships(ctx, tx); err != nil {
		return fmt.Errorf("failed to populate memberships: %w", err)
	}
	if err = p.populateEndDevices(ctx, tx); err != nil {
		return fmt.Errorf("failed to populate end devices: %w", err)
	}
	return nil
}

func (p *Populator) populateApplications(ctx context.Context, db *gorm.DB) (err error) {
	for i, application := range p.Applications {
		p.Applications[i], err = GetApplicationStore(db).CreateApplication(ctx, application)
		if err != nil {
			return err
		}
		p.Applications[i].ContactInfo, err = GetContactInfoStore(db).SetContactInfo(ctx, application, application.ContactInfo)
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
		p.Clients[i].ContactInfo, err = GetContactInfoStore(db).SetContactInfo(ctx, client, client.ContactInfo)
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
		p.Gateways[i].ContactInfo, err = GetContactInfoStore(db).SetContactInfo(ctx, gateway, gateway.ContactInfo)
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
		p.Organizations[i].ContactInfo, err = GetContactInfoStore(db).SetContactInfo(ctx, organization, organization.ContactInfo)
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
		p.Users[i].ContactInfo, err = GetContactInfoStore(db).SetContactInfo(ctx, user.GetIds(), user.ContactInfo)
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
			apiKey.Id = generatedID
			apiKey.Key = hashedKey
			if _, err = GetAPIKeyStore(db).CreateAPIKey(ctx, entityID, apiKey); err != nil {
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
				member.Ids,
				entityID,
				ttnpb.RightsFrom(member.Rights...),
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Populator) populateEndDevices(ctx context.Context, db *gorm.DB) (err error) {
	for i, endDevice := range p.EndDevices {
		p.EndDevices[i], err = GetEndDeviceStore(db).CreateEndDevice(ctx, endDevice)
		if err != nil {
			return err
		}
	}
	return nil
}
