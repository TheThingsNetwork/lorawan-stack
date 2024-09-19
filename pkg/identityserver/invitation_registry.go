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

package identityserver

import (
	"context"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/email/templates"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var evtCreateInvitation = events.Define(
	"invitation.create", "create invitation",
	events.WithAuthFromContext(),
	events.WithClientInfoFromContext(),
)

var errNoInviteRights = errors.DefinePermissionDenied(
	"no_invite_rights",
	"no rights for inviting users",
)

func (is *IdentityServer) sendInvitation(ctx context.Context, in *ttnpb.SendInvitationRequest) (*ttnpb.Invitation, error) {
	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	if !authInfo.GetUniversalRights().IncludesAll(ttnpb.Right_RIGHT_SEND_INVITES) {
		return nil, errNoInviteRights.New()
	}
	token, err := auth.GenerateKey(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	ttl := is.configFromContext(ctx).UserRegistration.Invitation.TokenTTL
	expires := now.Add(ttl)
	invitation := &ttnpb.Invitation{
		Email:     in.Email,
		Token:     token,
		ExpiresAt: timestamppb.New(expires),
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		invitation, err = st.CreateInvitation(ctx, invitation)
		return err
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtCreateInvitation.NewWithIdentifiersAndData(ctx, nil, invitation))
	go is.SendTemplateEmailToUsers(is.FromRequestContext(ctx), ttnpb.NotificationType_INVITATION, func(ctx context.Context, data email.TemplateData) (email.TemplateData, error) {
		return &templates.InvitationData{
			TemplateData:    data,
			SenderIds:       authInfo.GetEntityIdentifiers().GetUserIds(),
			InvitationToken: invitation.Token,
			TTL:             ttl,
		}, nil
	}, &ttnpb.User{PrimaryEmailAddress: in.Email})
	return invitation, nil
}

func (is *IdentityServer) listInvitations(ctx context.Context, req *ttnpb.ListInvitationsRequest) (invitations *ttnpb.Invitations, err error) {
	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	if !authInfo.GetUniversalRights().IncludesAll(ttnpb.Right_RIGHT_SEND_INVITES) {
		return nil, errNoInviteRights.New()
	}
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	invitations = &ttnpb.Invitations{}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		invitations.Invitations, err = st.FindInvitations(ctx)
		return err
	})
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

func (is *IdentityServer) deleteInvitation(ctx context.Context, in *ttnpb.DeleteInvitationRequest) (*emptypb.Empty, error) {
	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	if !authInfo.GetUniversalRights().IncludesAll(ttnpb.Right_RIGHT_SEND_INVITES) {
		return nil, errNoInviteRights.New()
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		return st.DeleteInvitation(ctx, in.Email)
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

type invitationRegistry struct {
	ttnpb.UnimplementedUserInvitationRegistryServer

	*IdentityServer
}

func (ir *invitationRegistry) Send(ctx context.Context, req *ttnpb.SendInvitationRequest) (*ttnpb.Invitation, error) {
	return ir.sendInvitation(ctx, req)
}

func (ir *invitationRegistry) List(ctx context.Context, req *ttnpb.ListInvitationsRequest) (*ttnpb.Invitations, error) {
	return ir.listInvitations(ctx, req)
}

func (ir *invitationRegistry) Delete(ctx context.Context, req *ttnpb.DeleteInvitationRequest) (*emptypb.Empty, error) {
	return ir.deleteInvitation(ctx, req)
}
