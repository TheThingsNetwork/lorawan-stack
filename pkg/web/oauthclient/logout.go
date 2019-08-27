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

package oauthclient

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// HandleLogout invalidates the user's authorization and removes the auth
// cookie.
func (oc *OAuthClient) HandleLogout(c echo.Context) error {
	token, err := oc.freshToken(c)
	if err != nil {
		return err
	}

	creds := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     token.AccessToken,
		AllowInsecure: oc.component.AllowInsecureForCredentials(),
	})

	ctx := c.Request().Context()

	if peer, err := oc.component.GetPeer(ctx, ttnpb.ClusterRole_ACCESS, nil); err == nil {
		if cc, err := peer.Conn(); err == nil {
			if res, err := ttnpb.NewEntityAccessClient(cc).AuthInfo(ctx, ttnpb.Empty, creds); err == nil {
				if tokenInfo := res.GetOAuthAccessToken(); tokenInfo != nil {
					_, err := ttnpb.NewOAuthAuthorizationRegistryClient(cc).DeleteToken(ctx, &ttnpb.OAuthAccessTokenIdentifiers{
						UserIDs:   tokenInfo.UserIDs,
						ClientIDs: tokenInfo.ClientIDs,
						ID:        tokenInfo.ID,
					}, creds)
					if err != nil {
						log.FromContext(ctx).WithError(err).Error("Could not invalidate access token")
					}
				}
			}
		}
	}

	oc.removeAuthCookie(c)
	return c.NoContent(http.StatusNoContent)
}
