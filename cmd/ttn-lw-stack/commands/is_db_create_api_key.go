// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver"
	store "go.thethings.network/lorawan-stack/v3/pkg/identityserver/gormstore"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	errExpiryDateInPast  = errors.DefineInvalidArgument("expiry_date_invalid", "expiry date is in the past")
	errInvalidDateFormat = errors.DefineInvalidArgument("expiry_date_format_invalid", "invalid expiry date format (RFC3339: YYYY-MM-DDTHH:MM:SSZ)")
)

var createAPIKeyCommand = &cobra.Command{
	Use:   "create-user-api-key",
	Short: "Create an API key with full rights on the user in the Identity Server database",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		logger.Info("Connecting to Identity Server database...")
		db, err := store.Open(ctx, config.IS.DatabaseURI)
		if err != nil {
			return err
		}
		defer db.Close()

		userID, err := cmd.Flags().GetString("user-id")
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")

		expiry, _ := cmd.Flags().GetString("api-key-expiry")
		var expiryDate *time.Time

		if expiry != "" {
			expiryDate, err := time.Parse(time.RFC3339, expiry)
			if err != nil {
				return errInvalidDateFormat.New()
			}
			if expiryDate.Before(time.Now()) {
				return errExpiryDateInPast.New()
			}
		}

		usr := &ttnpb.User{
			Ids: &ttnpb.UserIdentifiers{UserId: userID},
		}
		rights := []ttnpb.Right{ttnpb.Right_RIGHT_ALL}
		apiKeyStore := store.GetAPIKeyStore(db)
		key, token, err := is.GenerateAPIKey(ctx, name, expiryDate, rights...)
		if err != nil {
			return err
		}
		key, err = apiKeyStore.CreateAPIKey(ctx, usr.GetEntityIdentifiers(), key)
		if err != nil {
			return err
		}
		key.Key = token
		logger.Infof("API key ID: %s", key.Id)
		logger.Infof("API key value: %s", key.Key)
		logger.Warn("The API key value will never be shown again")
		logger.Warn("Make sure to copy it to a safe place")

		return io.Write(os.Stdout, config.OutputFormat, key)
	},
}

func init() {
	createAPIKeyCommand.Flags().String("user-id", "admin", "User ID")
	createAPIKeyCommand.Flags().String("name", "admin-api-key", "API key name")
	createAPIKeyCommand.Flags().String("api-key-expiry", "", "(YYYY-MM-DDTHH:MM:SSZ)")
	isDBCommand.AddCommand(createAPIKeyCommand)
}
