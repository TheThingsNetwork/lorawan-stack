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

package oauth

import (
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	// EvtUserLogin indicates a user login.
	EvtUserLogin = events.Define(
		"oauth.user.login", "login user successful",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_ALL),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	// EvtUserLogout indicates a user logout.
	EvtUserLogout = events.Define(
		"oauth.user.logout", "logout user",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_ALL),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)

	evtAuthorize = events.Define(
		"oauth.authorize", "authorize OAuth client",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_AUTHORIZED_CLIENTS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtTokenExchange = events.Define(
		"oauth.token.exchange", "exchange OAuth access token",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_AUTHORIZED_CLIENTS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtAccessTokenDeleted = events.Define(
		"oauth.token.deleted", "delete access token",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_AUTHORIZED_CLIENTS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUserSessionTerminated = events.Define(
		"oauth.session.terminated", "terminate user session",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_AUTHORIZED_CLIENTS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)
