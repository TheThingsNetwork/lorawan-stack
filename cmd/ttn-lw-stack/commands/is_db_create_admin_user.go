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
	"context"
	"os"
	"time"

	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	bunstore "go.thethings.network/lorawan-stack/v3/pkg/identityserver/bunstore"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	storeutil "go.thethings.network/lorawan-stack/v3/pkg/util/store"
)

var errPasswordMismatch = errors.DefineInvalidArgument("password_mismatch", "password did not match")

var createAdminUserCommand = &cobra.Command{
	Use:   "create-admin-user",
	Short: "Create an admin user in the Identity Server database",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		logger.Info("Connecting to Identity Server database...")
		db, err := storeutil.OpenDB(ctx, config.IS.DatabaseURI)
		if err != nil {
			return err
		}
		bunDB := bun.NewDB(db, pgdialect.New())
		st, err := bunstore.NewStore(ctx, bunDB)
		if err != nil {
			return err
		}
		defer db.Close()

		userID, err := cmd.Flags().GetString("id")
		if err != nil {
			return err
		}
		email, err := cmd.Flags().GetString("email")
		if err != nil {
			return err
		}
		if email == "" {
			return errMissingFlag.WithAttributes("flag", "email")
		}
		password, err := cmd.Flags().GetString("password")
		if err != nil {
			return err
		}
		if password == "" {
			pw, err := gopass.GetPasswdPrompt("Please enter user password:", true, os.Stdin, os.Stderr)
			if err != nil {
				return err
			}
			password = string(pw)
			pw, err = gopass.GetPasswdPrompt("Please repeat user password:", true, os.Stdin, os.Stderr)
			if err != nil {
				return err
			}
			if string(pw) != password {
				return errPasswordMismatch.New()
			}
		}
		if password == "" {
			return errMissingFlag.WithAttributes("flag", "password")
		}
		hashedPassword, err := auth.Hash(ctx, password)
		if err != nil {
			return err
		}

		now := time.Now()

		usrFieldMask := []string{
			"primary_email_address",
			"primary_email_address_validated_at",
			"password",
			"password_updated_at",
			"state",
			"admin",
		}
		usr := &ttnpb.User{
			Ids: &ttnpb.UserIdentifiers{UserId: userID},
		}

		var usrExists bool
		if _, err := st.GetUser(ctx, usr.GetIds(), usrFieldMask); err == nil {
			usrExists = true
		}
		usr.PrimaryEmailAddress = email
		usr.PrimaryEmailAddressValidatedAt = ttnpb.ProtoTimePtr(now)
		usr.Password = hashedPassword
		usr.PasswordUpdatedAt = ttnpb.ProtoTimePtr(now)
		usr.State = ttnpb.State_STATE_APPROVED
		usr.Admin = true

		if usrExists {
			logger.Info("Updating user...")
			if _, err = st.UpdateUser(ctx, usr, usrFieldMask); err != nil {
				return err
			}
			logger.Info("Updated user")
		} else {
			logger.Info("Creating user...")
			if _, err = st.CreateUser(ctx, usr); err != nil {
				return err
			}
			logger.Info("Created user")
		}

		return nil
	},
}

func init() {
	createAdminUserCommand.Flags().String("id", "admin", "User ID")
	createAdminUserCommand.Flags().String("email", "", "Email address")
	createAdminUserCommand.Flags().String("password", "", "Password")
	isDBCommand.AddCommand(createAdminUserCommand)
}
