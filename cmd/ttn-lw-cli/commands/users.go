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
	"os"

	"github.com/gogo/protobuf/types"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	selectUserFlags     = util.FieldMaskFlags(&ttnpb.User{})
	setUserFlags        = util.FieldFlags(&ttnpb.User{})
	profilePictureFlags = &pflag.FlagSet{}

	selectAllUserFlags = util.SelectAllFlagSet("user")
)

func userIDFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("user-id", "", "")
	return flagSet
}

var errNoUserID = errors.DefineInvalidArgument("no_user_id", "no user ID set")

func getUserID(flagSet *pflag.FlagSet, args []string) *ttnpb.UserIdentifiers {
	var userID string
	if len(args) > 0 {
		if len(args) > 1 {
			logger.Warn("Multiple IDs found in arguments, considering only the first")
		}
		userID = args[0]
	} else {
		userID, _ = flagSet.GetString("user-id")
	}
	if userID == "" {
		return nil
	}
	return &ttnpb.UserIdentifiers{UserID: userID}
}

func printPasswordRequirements(msg *ttnpb.IsConfiguration_UserRegistration_PasswordRequirements) {
	if msg == nil {
		return
	}
	logger.Info("Password Requirements:")
	if v := msg.GetMinLength().GetValue(); v > 0 {
		logger.Infof("Minimum length: %d", v)
	}
	if v := msg.GetMaxLength().GetValue(); v > 0 {
		logger.Infof("Maximum length: %d", v)
	}
	if v := msg.GetMinUppercase().GetValue(); v > 0 {
		logger.Infof("Minimum uppercase: %d", v)
	}
	if v := msg.GetMinDigits().GetValue(); v > 0 {
		logger.Infof("Minimum digits: %d", v)
	}
	if v := msg.GetMinSpecial().GetValue(); v > 0 {
		logger.Infof("Minimum special characters: %d", v)
	}
}

var errPasswordMismatch = errors.DefineInvalidArgument("password_mismatch", "password did not match")

var (
	usersCommand = &cobra.Command{
		Use:     "users",
		Aliases: []string{"user", "usr", "u"},
		Short:   "User commands",
	}
	usersListCommand = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List users",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := util.SelectFieldMask(cmd.Flags(), selectUserFlags)
			paths = ttnpb.AllowedFields(paths, ttnpb.AllowedFieldMaskPathsForRPC["/ttn.lorawan.v3.UserRegistry/List"])

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttnpb.NewUserRegistryClient(is).List(ctx, &ttnpb.ListUsersRequest{
				FieldMask: types.FieldMask{Paths: paths},
				Limit:     limit,
				Page:      page,
				Order:     getOrder(cmd.Flags()),
			}, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Users)
		},
	}
	usersSearchCommand = &cobra.Command{
		Use:   "search",
		Short: "Search for users",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := util.SelectFieldMask(cmd.Flags(), selectUserFlags)
			paths = ttnpb.AllowedFields(paths, ttnpb.AllowedFieldMaskPathsForRPC["/ttn.lorawan.v3.EntityRegistrySearch/SearchUsers"])

			req, opt, getTotal := getSearchEntitiesRequest(cmd.Flags())
			req.FieldMask.Paths = paths

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewEntityRegistrySearchClient(is).SearchUsers(ctx, req, opt)
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Users)
		},
	}
	usersGetCommand = &cobra.Command{
		Use:     "get [user-id]",
		Aliases: []string{"info"},
		Short:   "Get a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			usrID := getUserID(cmd.Flags(), args)
			if usrID == nil {
				return errNoUserID
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectUserFlags)
			paths = ttnpb.AllowedFields(paths, ttnpb.AllowedFieldMaskPathsForRPC["/ttn.lorawan.v3.UserRegistry/Get"])

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewUserRegistryClient(is).Get(ctx, &ttnpb.GetUserRequest{
				UserIdentifiers: *usrID,
				FieldMask:       types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	usersCreateCommand = &cobra.Command{
		Use:               "create [user-id]",
		Aliases:           []string{"add", "register"},
		Short:             "Create a user",
		PersistentPreRunE: preRun(optionalAuth),
		RunE: asBulk(func(cmd *cobra.Command, args []string) (err error) {
			usrID := getUserID(cmd.Flags(), args)
			var user ttnpb.User
			user.State = ttnpb.STATE_APPROVED // This may not be honored by the server.
			if inputDecoder != nil {
				_, err := inputDecoder.Decode(&user)
				if err != nil {
					return err
				}
			}
			if err := util.SetFields(&user, setUserFlags); err != nil {
				return err
			}
			user.Attributes = mergeAttributes(user.Attributes, cmd.Flags())
			if usrID != nil && usrID.UserID != "" {
				user.UserID = usrID.UserID
			}
			if user.UserID == "" {
				return errNoUserID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}

			isConfig, err := ttnpb.NewIsClient(is).GetConfiguration(ctx, &ttnpb.GetIsConfigurationRequest{})
			if err != nil {
				if errors.IsUnimplemented(err) {
					logger.Warn("Could not get user registration configuration from the server, but we can continue without it")
				} else {
					return err
				}
			}
			registrationConfig := isConfig.GetConfiguration().GetUserRegistration()

			if user.Password == "" {
				printPasswordRequirements(registrationConfig.GetPasswordRequirements())
				pw, err := gopass.GetPasswdPrompt("Please enter password:", true, os.Stdin, os.Stderr)
				if err != nil {
					return err
				}
				user.Password = string(pw)
				pw, err = gopass.GetPasswdPrompt("Please confirm password:", true, os.Stdin, os.Stderr)
				if err != nil {
					return err
				}
				if string(pw) != user.Password {
					return errPasswordMismatch
				}
			}

			if !isConfig.GetConfiguration().GetProfilePicture().GetDisableUpload().GetValue() {
				if profilePicture, err := cmd.Flags().GetString("profile_picture"); err == nil && profilePicture != "" {
					user.ProfilePicture, err = readPicture(profilePicture)
					if err != nil {
						return err
					}
				}
			}

			invitationToken, _ := cmd.Flags().GetString("invitation-token")

			res, err := ttnpb.NewUserRegistryClient(is).Create(ctx, &ttnpb.CreateUserRequest{
				User:            user,
				InvitationToken: invitationToken,
			})
			if err != nil {
				return err
			}

			if err = io.Write(os.Stdout, config.OutputFormat, res); err != nil {
				return err
			}

			if registrationConfig.GetContactInfoValidation().GetRequired().GetValue() {
				logger.Warn("You need to confirm your email address before you can use your user account")
			}
			if registrationConfig.GetAdminApproval().GetRequired().GetValue() {
				logger.Warn("Your user account needs to be approved by an admin before you can use it")
			}

			return nil
		}),
	}
	usersSetCommand = &cobra.Command{
		Use:     "set [user-id]",
		Aliases: []string{"update"},
		Short:   "Set properties of a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			usrID := getUserID(cmd.Flags(), args)
			if usrID == nil {
				return errNoUserID
			}
			paths := util.UpdateFieldMask(cmd.Flags(), setUserFlags, attributesFlags(), profilePictureFlags)
			if len(paths) == 0 {
				logger.Warn("No fields selected, won't update anything")
				return nil
			}
			if ttnpb.ContainsField("password", paths) {
				logger.Warn("Most servers do not allow changing the password of a user like this.")
				logger.Warnf("Use \"%s [user-id]\" to change your password.", usersUpdatePasswordCommand.CommandPath())
				logger.Warnf("Use \"%s [user-id]\" if you forgot your password and want to request a temporary password.", usersForgotPasswordCommand.CommandPath())
				logger.Warnf("Alternatively, admins may use \"%s [user-id] --temporary-password [pass]\" to set a temporary password for a user.", cmd.CommandPath())
				logger.Warn("The user can then use this temporary password when changing their password.")
			}
			var user ttnpb.User
			if err := util.SetFields(&user, setUserFlags); err != nil {
				return err
			}
			user.Attributes = mergeAttributes(user.Attributes, cmd.Flags())
			user.UserIdentifiers = *usrID
			if profilePicture, err := cmd.Flags().GetString("profile_picture"); err == nil && profilePicture != "" {
				user.ProfilePicture, err = readPicture(profilePicture)
				if err != nil {
					return err
				}
				paths = append(paths, "profile_picture")
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewUserRegistryClient(is).Update(ctx, &ttnpb.UpdateUserRequest{
				User:      user,
				FieldMask: types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			res.SetFields(&user, "ids")
			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	usersForgotPasswordCommand = &cobra.Command{
		Use:               "forgot-password [user-id]",
		Short:             "Request a temporary user password",
		PersistentPreRunE: preRun(),
		RunE: func(cmd *cobra.Command, args []string) error {
			usrID := getUserID(cmd.Flags(), args)
			if usrID == nil {
				return errNoUserID
			}

			usrID.Email, _ = cmd.Flags().GetString("email")

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewUserRegistryClient(is).CreateTemporaryPassword(ctx, &ttnpb.CreateTemporaryPasswordRequest{
				UserIdentifiers: *usrID,
			})

			return err
		},
	}
	usersUpdatePasswordCommand = &cobra.Command{
		Use:               "update-password [user-id]",
		Aliases:           []string{"change-password"},
		Short:             "Update a user password",
		PersistentPreRunE: preRun(),
		RunE: func(cmd *cobra.Command, args []string) error {
			refreshToken() // NOTE: ignore errors.
			optionalAuth()

			usrID := getUserID(cmd.Flags(), args)
			if usrID == nil {
				return errNoUserID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}

			isConfig, err := ttnpb.NewIsClient(is).GetConfiguration(ctx, &ttnpb.GetIsConfigurationRequest{})
			if err != nil {
				if errors.IsUnimplemented(err) {
					logger.Warn("Could not get password requirements from the server, but we can continue without knowing those")
				} else {
					return err
				}
			}

			old, _ := cmd.Flags().GetString("old")
			if old == "" {
				pw, err := gopass.GetPasswdPrompt("Please enter old password:", true, os.Stdin, os.Stderr)
				if err != nil {
					return err
				}
				old = string(pw)
			}

			new, _ := cmd.Flags().GetString("new")
			if new == "" {
				printPasswordRequirements(isConfig.GetConfiguration().GetUserRegistration().GetPasswordRequirements())
				pw, err := gopass.GetPasswdPrompt("Please enter new password:", true, os.Stdin, os.Stderr)
				if err != nil {
					return err
				}
				new = string(pw)
			}

			revokeAllAccess, _ := cmd.Flags().GetBool("revoke-all-access")

			res, err := ttnpb.NewUserRegistryClient(is).UpdatePassword(ctx, &ttnpb.UpdateUserPasswordRequest{
				UserIdentifiers: *usrID,
				Old:             old,
				New:             new,
				RevokeAllAccess: revokeAllAccess,
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	usersDeleteCommand = &cobra.Command{
		Use:     "delete [user-id]",
		Aliases: []string{"del", "remove", "rm"},
		Short:   "Delete a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			usrID := getUserID(cmd.Flags(), args)
			if usrID == nil {
				return errNoUserID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewUserRegistryClient(is).Delete(ctx, usrID)
			if err != nil {
				return err
			}

			return nil
		},
	}
	usersContactInfoCommand = contactInfoCommands("user", func(cmd *cobra.Command, args []string) (*ttnpb.EntityIdentifiers, error) {
		usrID := getUserID(cmd.Flags(), args)
		if usrID == nil {
			return nil, errNoUserID
		}
		return usrID.EntityIdentifiers(), nil
	})
)

func init() {
	profilePictureFlags.String("profile_picture", "", "upload the profile picture from this file")

	usersListCommand.Flags().AddFlagSet(selectUserFlags)
	usersListCommand.Flags().AddFlagSet(selectAllUserFlags)
	usersListCommand.Flags().AddFlagSet(paginationFlags())
	usersListCommand.Flags().AddFlagSet(orderFlags())
	usersCommand.AddCommand(usersListCommand)
	usersSearchCommand.Flags().AddFlagSet(searchFlags())
	usersSearchCommand.Flags().AddFlagSet(selectAllUserFlags)
	usersSearchCommand.Flags().AddFlagSet(selectUserFlags)
	usersCommand.AddCommand(usersSearchCommand)
	usersGetCommand.Flags().AddFlagSet(userIDFlags())
	usersGetCommand.Flags().AddFlagSet(selectUserFlags)
	usersGetCommand.Flags().AddFlagSet(selectAllUserFlags)
	usersCommand.AddCommand(usersGetCommand)
	usersCreateCommand.Flags().AddFlagSet(userIDFlags())
	usersCreateCommand.Flags().AddFlagSet(setUserFlags)
	usersCreateCommand.Flags().AddFlagSet(attributesFlags())
	usersCreateCommand.Flags().AddFlagSet(profilePictureFlags)
	usersCreateCommand.Flags().Lookup("state").DefValue = ttnpb.STATE_APPROVED.String()
	usersCreateCommand.Flags().String("invitation-token", "", "")
	usersCommand.AddCommand(usersCreateCommand)
	usersSetCommand.Flags().AddFlagSet(userIDFlags())
	usersSetCommand.Flags().AddFlagSet(setUserFlags)
	usersSetCommand.Flags().AddFlagSet(attributesFlags())
	usersSetCommand.Flags().AddFlagSet(profilePictureFlags)
	usersCommand.AddCommand(usersSetCommand)
	usersForgotPasswordCommand.Flags().AddFlagSet(userIDFlags())
	usersForgotPasswordCommand.Flags().String("email", "", "")
	usersCommand.AddCommand(usersForgotPasswordCommand)
	usersUpdatePasswordCommand.Flags().AddFlagSet(userIDFlags())
	usersUpdatePasswordCommand.Flags().String("old", "", "")
	usersUpdatePasswordCommand.Flags().String("new", "", "")
	usersUpdatePasswordCommand.Flags().Bool("revoke-all-access", true, "revoke all sessions and access tokens after the password is updated")
	usersCommand.AddCommand(usersUpdatePasswordCommand)
	usersDeleteCommand.Flags().AddFlagSet(userIDFlags())
	usersCommand.AddCommand(usersDeleteCommand)
	usersContactInfoCommand.PersistentFlags().AddFlagSet(userIDFlags())
	usersCommand.AddCommand(usersContactInfoCommand)
	Root.AddCommand(usersCommand)
}
