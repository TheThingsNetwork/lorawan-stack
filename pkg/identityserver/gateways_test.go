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

package identityserver

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc/metadata"
)

const newConfigurationTimeout = 5 * time.Second

var _ ttnpb.IsGatewayServer = new(gatewayService)

func TestGatewaysBlacklistedIDs(t *testing.T) {
	a := assertions.New(t)
	is := newTestIS(t)

	ctx := newTestCtx(newTestUsers()["bob"].UserIdentifiers)

	// Can not create gateways with blacklisted IDs.
	for _, id := range newTestSettings().BlacklistedIDs {
		_, err := is.gatewayService.CreateGateway(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: id},
			},
		})
		a.So(err, should.NotBeNil)
		a.So(err, should.DescribeError, ErrBlacklistedID)
	}
}

func TestGateways(t *testing.T) {
	for _, tc := range []struct {
		tcname string
		gids   ttnpb.GatewayIdentifiers
		sids   ttnpb.GatewayIdentifiers
	}{
		{
			"SearchByGatewayID",
			ttnpb.GatewayIdentifiers{
				GatewayID: "foo-gtw",
				EUI:       &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
			},
			ttnpb.GatewayIdentifiers{
				GatewayID: "foo-gtw",
			},
		},
		{
			"SearchByEUI",
			ttnpb.GatewayIdentifiers{
				GatewayID: "foo-gtw",
				EUI:       &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
			},
			ttnpb.GatewayIdentifiers{
				EUI: &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
			},
		},
		{
			"SearchByAllIdentifiers",
			ttnpb.GatewayIdentifiers{
				GatewayID: "foo-gtw",
				EUI:       &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
			},
			ttnpb.GatewayIdentifiers{
				GatewayID: "foo-gtw",
				EUI:       &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
			},
		},
	} {
		t.Run(tc.tcname, func(t *testing.T) {
			testGateways(t, tc.gids, tc.sids)
		})
	}
}

type pullConfigurationServer struct {
	test.MockServerStream

	gtwConfigs chan *ttnpb.Gateway
	ctx        context.Context
}

func (s *pullConfigurationServer) Send(gtw *ttnpb.Gateway) error {
	s.gtwConfigs <- gtw
	return nil
}

func newPullConfigurationServer() *pullConfigurationServer {
	srv := &pullConfigurationServer{
		gtwConfigs: make(chan *ttnpb.Gateway),
	}
	srv.MockServerStream = test.MockServerStream{
		MockStream: &test.MockStream{
			ContextFunc: func() context.Context {
				return srv.ctx
			},
		},
	}

	return srv
}

func TestPullConfiguration(t *testing.T) {
	is := newTestIS(t)

	user := newTestUsers()["bob"]
	userCtx := newTestCtx(user.UserIdentifiers)

	oldCooldown := updateDebounce
	updateDebounce = time.Duration(0)
	defer func() {
		updateDebounce = oldCooldown
	}()

	for _, tc := range []struct {
		name  string
		paths []string

		afterUpdateReception func(a *assertions.Assertion, sent, received *ttnpb.Gateway)
	}{
		{
			name:  "EmptyFieldPath",
			paths: []string{},
			afterUpdateReception: func(a *assertions.Assertion, _, received *ttnpb.Gateway) {
				a.So(received.GetAntennas(), should.HaveLength, 0)
				a.So(received.GetDisableTxDelay(), should.Equal, false)
				a.So(received.GetPlatform(), should.Equal, "")
				a.So(received.GetFrequencyPlanID(), should.Equal, "")
			},
		},
		{
			name:  "PopulatedFieldPath",
			paths: []string{"platform"},
			afterUpdateReception: func(a *assertions.Assertion, sent, received *ttnpb.Gateway) {
				a.So(received.GetDisableTxDelay(), should.BeFalse)
				a.So(received.GetFrequencyPlanID(), should.Equal, "")
				a.So(sent.GetPlatform(), should.Equal, received.GetPlatform())
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := assertions.New(t)

			gtw := *ttnpb.NewPopulatedGateway(test.Randy, false)
			_, err := is.gatewayService.CreateGateway(userCtx, &ttnpb.CreateGatewayRequest{
				Gateway: gtw,
			})
			a.So(err, should.BeNil)

			apiKeyRequest := &ttnpb.GenerateGatewayAPIKeyRequest{
				Name:               tc.name,
				GatewayIdentifiers: gtw.GatewayIdentifiers,
				Rights:             []ttnpb.Right{ttnpb.RIGHT_GATEWAY_INFO, ttnpb.RIGHT_GATEWAY_LINK},
			}
			key, err := is.userService.GenerateGatewayAPIKey(userCtx, apiKeyRequest)
			a.So(err, should.BeNil)

			gtwCtx := test.Context()
			md := metadata.New(map[string]string{
				"id":            unique.ID(gtwCtx, gtw.GatewayIdentifiers),
				"authorization": fmt.Sprintf("Bearer %s", key.GetKey()),
			})
			if ctxMd, ok := metadata.FromIncomingContext(gtwCtx); ok {
				md = metadata.Join(ctxMd, md)
			}
			gtwCtx = metadata.NewIncomingContext(gtwCtx, md)

			wg := sync.WaitGroup{}
			wg.Add(1)

			stream := newPullConfigurationServer()
			pullConfigCtx, cancel := context.WithCancel(gtwCtx)
			stream.ctx = pullConfigCtx

			go func() {
				err := is.gatewayService.PullConfiguration(&ttnpb.PullConfigurationRequest{
					GatewayIdentifiers: gtw.GatewayIdentifiers,
					ProjectionMask: &pbtypes.FieldMask{
						Paths: tc.paths,
					},
				}, stream)
				a.So(err, should.Equal, context.Canceled)
				wg.Done()
			}()

			select {
			case cfg := <-stream.gtwConfigs:
				tc.afterUpdateReception(a, &gtw, cfg)
			case <-time.After(5 * time.Second):
				t.Fatal("Did not receive initial config after 5 seconds")
			}

			updatedGtw := *ttnpb.NewPopulatedGateway(test.Randy, false)
			updatedGtw.GatewayIdentifiers = gtw.GatewayIdentifiers
			_, err = is.gatewayService.UpdateGateway(userCtx, &ttnpb.UpdateGatewayRequest{
				Gateway: updatedGtw,
				UpdateMask: pbtypes.FieldMask{
					Paths: []string{"antennas", "disable_tx_delay", "platform", "frequency_plan_id"},
				},
			})
			a.So(err, should.BeNil)

			select {
			case received := <-stream.gtwConfigs:
				tc.afterUpdateReception(a, &updatedGtw, received)
			case <-time.After(5 * time.Second):
				t.Fatal("Did not receive updated config after 5 seconds")
			}

			cancel()

			wg.Wait()
		})
	}
}

func testGateways(t *testing.T, gids, sids ttnpb.GatewayIdentifiers) {
	a := assertions.New(t)
	is := newTestIS(t)

	user := newTestUsers()["bob"]

	gtw := ttnpb.Gateway{
		GatewayIdentifiers: gids,
		ClusterAddress:     "localhost:1234",
		FrequencyPlanID:    "868.8",
		Attributes: map[string]string{
			"version": "1.2",
		},
		Antennas: []ttnpb.GatewayAntenna{
			{
				Gain: 1.1,
				Location: ttnpb.Location{
					Latitude:  1.1,
					Longitude: 1.1,
				},
			},
			{
				Gain: 2.2,
				Location: ttnpb.Location{
					Latitude:  2.2,
					Longitude: 2.2,
				},
			},
			{
				Gain: 3,
				Location: ttnpb.Location{
					Latitude:  3,
					Longitude: 3,
				},
			},
		},
		Radios:         []ttnpb.GatewayRadio{},
		DisableTxDelay: true,
	}

	ctx := newTestCtx(user.UserIdentifiers)

	_, err := is.gatewayService.CreateGateway(ctx, &ttnpb.CreateGatewayRequest{
		Gateway: gtw,
	})
	a.So(err, should.BeNil)

	found, err := is.gatewayService.GetGateway(ctx, &sids)
	a.So(err, should.BeNil)
	a.So(found, should.EqualFieldsWithIgnores(GatewayGeneratedFields...), gtw)

	gtws, err := is.gatewayService.ListGateways(ctx, &ttnpb.ListGatewaysRequest{})
	a.So(err, should.BeNil)
	if a.So(gtws.Gateways, should.HaveLength, 1) {
		a.So(gtws.Gateways[0], should.EqualFieldsWithIgnores(GatewayGeneratedFields...), gtw)
	}

	gtw.Description = "foo"
	_, err = is.gatewayService.UpdateGateway(ctx, &ttnpb.UpdateGatewayRequest{
		Gateway: gtw,
		UpdateMask: pbtypes.FieldMask{
			Paths: []string{"description"},
		},
	})
	a.So(err, should.BeNil)

	found, err = is.gatewayService.GetGateway(ctx, &sids)
	a.So(err, should.BeNil)
	a.So(found, should.EqualFieldsWithIgnores(GatewayGeneratedFields...), gtw)

	// Generate a new API key.
	key, err := is.gatewayService.GenerateGatewayAPIKey(ctx, &ttnpb.GenerateGatewayAPIKeyRequest{
		GatewayIdentifiers: sids,
		Name:               "foo",
		Rights:             ttnpb.AllGatewayRights(),
	})
	a.So(err, should.BeNil)
	a.So(key.Key, should.NotBeEmpty)
	a.So(key.Name, should.Equal, key.Name)
	a.So(key.Rights, should.Resemble, ttnpb.AllGatewayRights())

	// Update API key.
	key.Rights = []ttnpb.Right{ttnpb.Right(10)}
	_, err = is.gatewayService.UpdateGatewayAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
		GatewayIdentifiers: sids,
		Name:               key.Name,
		Rights:             key.Rights,
	})
	a.So(err, should.BeNil)

	// Can not generate another API Key with the same name.
	_, err = is.gatewayService.GenerateGatewayAPIKey(ctx, &ttnpb.GenerateGatewayAPIKeyRequest{
		GatewayIdentifiers: sids,
		Name:               key.Name,
		Rights:             []ttnpb.Right{ttnpb.Right(1)},
	})
	a.So(err, should.NotBeNil)
	a.So(err, should.DescribeError, store.ErrAPIKeyNameConflict)

	keys, err := is.gatewayService.ListGatewayAPIKeys(ctx, &sids)
	a.So(err, should.BeNil)
	if a.So(keys.APIKeys, should.HaveLength, 1) {
		sort.Slice(keys.APIKeys[0].Rights, func(i, j int) bool { return keys.APIKeys[0].Rights[i] < keys.APIKeys[0].Rights[j] })
		a.So(keys.APIKeys[0], should.Resemble, key)
	}

	_, err = is.gatewayService.RemoveGatewayAPIKey(ctx, &ttnpb.RemoveGatewayAPIKeyRequest{
		GatewayIdentifiers: sids,
		Name:               key.Name,
	})
	a.So(err, should.BeNil)

	keys, err = is.gatewayService.ListGatewayAPIKeys(ctx, &sids)
	a.So(err, should.BeNil)
	a.So(keys.APIKeys, should.HaveLength, 0)

	// Set a new collaborator with SETTINGS_COLLABORATORS and INFO rights.
	alice := newTestUsers()["alice"]
	collab := &ttnpb.GatewayCollaborator{
		OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &alice.UserIdentifiers}},
		GatewayIdentifiers:            sids,
		Rights:                        []ttnpb.Right{ttnpb.RIGHT_GATEWAY_INFO, ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS},
	}

	_, err = is.gatewayService.SetGatewayCollaborator(ctx, collab)
	a.So(err, should.BeNil)

	rights, err := is.gatewayService.ListGatewayRights(ctx, &sids)
	a.So(err, should.BeNil)
	a.So(rights.Rights, should.Resemble, ttnpb.AllGatewayRights())

	collabs, err := is.gatewayService.ListGatewayCollaborators(ctx, &sids)
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 2)
	a.So(collabs.Collaborators, should.Contain, collab)
	a.So(collabs.Collaborators, should.Contain, &ttnpb.GatewayCollaborator{
		OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &user.UserIdentifiers}},
		GatewayIdentifiers:            sids,
		Rights:                        ttnpb.AllGatewayRights(),
	})

	// The new collaborator can not grant himself more rights.
	{
		collab.Rights = append(collab.Rights, ttnpb.RIGHT_GATEWAY_SETTINGS_API_KEYS)

		ctx := newTestCtx(alice.UserIdentifiers)

		_, err = is.gatewayService.SetGatewayCollaborator(ctx, collab)
		a.So(err, should.BeNil)

		rights, err := is.gatewayService.ListGatewayRights(ctx, &sids)
		a.So(err, should.BeNil)
		a.So(rights.Rights, should.HaveLength, 2)
		a.So(rights.Rights, should.NotContain, ttnpb.RIGHT_GATEWAY_SETTINGS_API_KEYS)

		// But they can revoke themselves the INFO right.
		collab.Rights = []ttnpb.Right{ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS}
		_, err = is.gatewayService.SetGatewayCollaborator(ctx, collab)
		a.So(err, should.BeNil)

		rights, err = is.gatewayService.ListGatewayRights(ctx, &sids)
		a.So(err, should.BeNil)
		a.So(rights.Rights, should.HaveLength, 1)
		a.So(rights.Rights, should.NotContain, ttnpb.RIGHT_GATEWAY_INFO)
	}

	// Trying to unset the main collaborator will result in an error as the gateway
	// will become unmanageable.
	_, err = is.gatewayService.SetGatewayCollaborator(ctx, &ttnpb.GatewayCollaborator{
		GatewayIdentifiers:            sids,
		OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &user.UserIdentifiers}},
	})
	a.So(err, should.NotBeNil)
	a.So(err, should.DescribeError, ErrUnmanageableGateway)

	// But we can revoke a shared right between the two collaborators.
	_, err = is.gatewayService.SetGatewayCollaborator(ctx, &ttnpb.GatewayCollaborator{
		GatewayIdentifiers:            sids,
		OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &user.UserIdentifiers}},
		Rights: ttnpb.DifferenceRights(ttnpb.AllGatewayRights(), []ttnpb.Right{ttnpb.RIGHT_GATEWAY_INFO}),
	})
	a.So(err, should.NotBeNil)

	collabs, err = is.gatewayService.ListGatewayCollaborators(ctx, &sids)
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 2)

	// Unset the last added collaborator.
	collab.Rights = []ttnpb.Right{}
	_, err = is.gatewayService.SetGatewayCollaborator(ctx, collab)
	a.So(err, should.BeNil)

	collabs, err = is.gatewayService.ListGatewayCollaborators(ctx, &sids)
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 1)

	_, err = is.gatewayService.DeleteGateway(ctx, &sids)
	a.So(err, should.BeNil)
}

type dummyEvent struct {
	t *testing.T

	name        string
	data        interface{}
	identifiers *ttnpb.CombinedIdentifiers
}

func (d dummyEvent) Context() context.Context {
	return log.NewContext(test.Context(), test.GetLogger(d.t))
}

func (d dummyEvent) Name() string {
	return d.name
}

func (d dummyEvent) Identifiers() *ttnpb.CombinedIdentifiers {
	return d.identifiers
}

func (d dummyEvent) Data() interface{} {
	return d.data
}

func (d dummyEvent) Time() time.Time          { panic("not implemented") }
func (d dummyEvent) CorrelationIDs() []string { panic("not implemented") }
func (d dummyEvent) Origin() string           { panic("not implemented") }
func (d dummyEvent) Caller() string           { panic("not implemented") }

func TestUpdateNotification(t *testing.T) {
	a := assertions.New(t)

	s := &gatewayService{
		IdentifiersFilter: events.NewIdentifierFilter(),
	}
	gtwIDs := &ttnpb.GatewayIdentifiers{
		GatewayID: "hello",
	}
	subscription := make(events.Channel, 1)
	s.IdentifiersFilter.Subscribe(test.Context(), gtwIDs, subscription)

	// Sending []string
	for _, data := range []interface{}{
		[]string{"platform"},
	} {
		go s.IdentifiersFilter.Notify(dummyEvent{
			t:    t,
			name: "is.gateway.update",
			data: data,
			identifiers: &ttnpb.CombinedIdentifiers{
				GatewayIDs: []*ttnpb.GatewayIdentifiers{gtwIDs},
			},
		})
		select {
		case updatedFields := <-subscription:
			a.So(updatedFields.Data().([]string), should.Contain, "platform")
		case <-time.After(newConfigurationTimeout):
			t.Fatal("No new subscription received")
		}
	}
}
