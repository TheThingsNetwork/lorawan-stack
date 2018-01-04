// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"context"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/claims"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestClient(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)

	user := testUsers()["bob"]

	cli := ttnpb.Client{
		ClientIdentifier: ttnpb.ClientIdentifier{"test-client"},
		Description:      "description foobarbaz",
		RedirectURI:      "foo.local/oauth",
		Secret:           "bar",
		Grants:           []ttnpb.GrantType{ttnpb.GRANT_REFRESH_TOKEN},
		Rights:           []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
		State:            ttnpb.STATE_PENDING,
		OfficialLabeled:  false,
		Creator:          ttnpb.UserIdentifier{user.UserID},
	}

	ctx := claims.NewContext(context.Background(), &auth.Claims{
		EntityID:  user.UserID,
		EntityTyp: auth.EntityUser,
		Source:    auth.Token,
		Rights:    append(ttnpb.AllUserRights),
	})

	_, err := is.CreateClient(ctx, &ttnpb.CreateClientRequest{
		Client: cli,
	})
	a.So(err, should.BeNil)

	// can't create clients with blacklisted ids
	for _, id := range testSettings().BlacklistedIDs {
		_, err := is.CreateClient(ctx, &ttnpb.CreateClientRequest{
			Client: ttnpb.Client{
				ClientIdentifier: ttnpb.ClientIdentifier{id},
			},
		})
		a.So(err, should.NotBeNil)
		a.So(ErrBlacklistedID.Describes(err), should.BeTrue)
	}

	found, err := is.GetClient(ctx, &ttnpb.ClientIdentifier{cli.ClientID})
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeClientIgnoringAutoFields, cli)

	clients, err := is.ListClients(ctx, &pbtypes.Empty{})
	a.So(err, should.BeNil)
	if a.So(clients.Clients, should.HaveLength, 1) {
		a.So(clients.Clients[0], test.ShouldBeClientIgnoringAutoFields, cli)
	}

	cli.Description = "foo"
	_, err = is.UpdateClient(ctx, &ttnpb.UpdateClientRequest{
		Client: cli,
		UpdateMask: pbtypes.FieldMask{
			Paths: []string{"description"},
		},
	})
	a.So(err, should.BeNil)

	found, err = is.GetClient(ctx, &ttnpb.ClientIdentifier{cli.ClientID})
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeClientIgnoringAutoFields, cli)

	_, err = is.DeleteClient(ctx, &ttnpb.ClientIdentifier{cli.ClientID})
	a.So(err, should.BeNil)

	found, err = is.GetClient(ctx, &ttnpb.ClientIdentifier{cli.ClientID})
	a.So(found, should.BeNil)
	a.So(err, should.NotBeNil)
	a.So(sql.ErrClientNotFound.Describes(err), should.BeTrue)
}
