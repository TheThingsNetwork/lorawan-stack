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

package commands

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"golang.org/x/oauth2"
)

var (
	loginCommand = &cobra.Command{
		Use:               "login",
		Short:             "Login",
		PersistentPreRunE: preRun(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, done := context.WithCancel(ctx)
			defer done()

			callback, err := cmd.Flags().GetBool("callback")
			if err != nil {
				return err
			}

			var token *oauth2.Token

			if callback {
				oauth2Config.RedirectURL = "http://localhost:11885/oauth/callback"

				http.HandleFunc("/oauth/callback", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodGet {
						http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
						return
					}
					token, err = oauth2Config.Exchange(ctx, r.URL.Query().Get("code"))
					if err != nil {
						msg := "Could not exchange OAuth access token"
						logger.WithError(err).Error(msg)
						w.WriteHeader(http.StatusUnauthorized)
						w.Write([]byte(msg))
						return
					}
					msg := "Got OAuth access token"
					logger.Info(msg)
					w.Write([]byte(msg))
					done()
				}))

				lis, err := net.Listen("tcp", ":11885")
				if err != nil {
					logger.WithError(err).Error("Could not listen for OAuth callback. Try re-running this command with --callback=false")
					return err
				}
				defer lis.Close()
				go http.Serve(lis, nil)
			} else {
				oauth2Config.RedirectURL = "code"
			}

			logger.Infof("Please go to %s", oauth2Config.AuthCodeURL(""))

			if callback {
				logger.Info("Waiting for your authorization...")
				<-ctx.Done()
			} else {
				var code string
				for {
					logger.Info("Please paste the authorization code and press enter")
					fmt.Fprint(os.Stderr, "> ")
					r := bufio.NewReader(os.Stdin)
					code, err = r.ReadString('\n')
					if err != nil {
						return err
					}
					code = strings.TrimSpace(code)
					tokenType, _, _, err := auth.SplitToken(code)
					if err != nil {
						logger.WithError(err).Warn("Could not parse authorization code")
						continue
					}
					if tokenType != auth.AuthorizationCode {
						logger.Warnf("Authorization codes should start with %s", auth.AuthorizationCode)
						continue
					}
					break
				}
				token, err = oauth2Config.Exchange(ctx, code)
				if err != nil {
					logger.WithError(err).Error("Could not exchange OAuth access token")
					return err
				}
				logger.Info("Got OAuth access token")
			}

			cache.Set("oauth_token", token)

			return nil
		},
	}
	logoutCommand = &cobra.Command{
		Use:   "logout",
		Short: "Logout",
		RunE: func(cmd *cobra.Command, args []string) error {
			if token, ok := cache.Get("oauth_token").(*oauth2.Token); ok && token != nil {
				is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
				if err != nil {
					return err
				}

				if res, err := ttnpb.NewEntityAccessClient(is).AuthInfo(ctx, ttnpb.Empty); err == nil {
					if tokenInfo := res.GetOAuthAccessToken(); tokenInfo != nil {
						_, err := ttnpb.NewOAuthAuthorizationRegistryClient(is).DeleteToken(ctx, &ttnpb.OAuthAccessTokenIdentifiers{
							UserIDs:   tokenInfo.UserIDs,
							ClientIDs: tokenInfo.ClientIDs,
							ID:        tokenInfo.ID,
						})
						if err != nil {
							logger.Warn("We could not revoke the OAuth token on the server")
							if time.Until(token.Expiry) > 0 {
								logger.Warnf("The OAuth token expires at %s", token.Expiry.Truncate(time.Minute).Format(time.Kitchen))
							}
							if token.RefreshToken != "" {
								logger.Warn("The OAuth token can still be refreshed after expiry")
							}
							logger.Warn("Please contact support if this token was compromised")
						}
					}
				}

				cache.Set("oauth_token", (*oauth2.Token)(nil))

				logger.Info("Logged out")
			}
			if _, ok := cache.Get("api_key").(string); ok {
				cache.Set("api_key", "")
				logger.Info("Removed API key from cache")
				logger.Warn("Make sure to delete the API key if it was compromised")
			}
			return nil
		},
	}
)

func init() {
	loginCommand.Flags().Bool("callback", true, "use local OAuth callback endpoint")
	Root.AddCommand(loginCommand)
	Root.AddCommand(logoutCommand)
}
