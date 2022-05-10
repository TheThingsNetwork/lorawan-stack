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
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	selectApplicationActivationSettingsFlags = util.NormalizedFlagSet()

	selectAllApplicationActivationSettingsFlags = util.SelectAllFlagSet("application activation settings")
)

var (
	applicationActivationSettingsCommand = &cobra.Command{
		Use:   "activation-settings",
		Short: "Application activation settings commands",
	}
	applicationActivationSettingsGetCommand = &cobra.Command{
		Use:   "get [application-id]",
		Short: "Get application activation settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID.New()
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectApplicationActivationSettingsFlags)
			if len(paths) == 0 {
				logger.Warn("No fields selected, will select everything")
				selectApplicationActivationSettingsFlags.VisitAll(func(flag *pflag.Flag) {
					paths = append(paths, strings.Replace(flag.Name, "-", "_", -1))
				})
			}
			paths = ttnpb.AllowedFields(paths, ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.ApplicationActivationSettingRegistry/Get"].Allowed)

			js, err := api.Dial(ctx, config.JoinServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationActivationSettingRegistryClient(js).Get(ctx, &ttnpb.GetApplicationActivationSettingsRequest{
				ApplicationIds: appID,
				FieldMask:      ttnpb.FieldMask(paths...),
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationActivationSettingsSetCommand = &cobra.Command{
		Use:     "set [application-id]",
		Aliases: []string{"update"},
		Short:   "Set application activation settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID.New()
			}

			var aas ttnpb.ApplicationActivationSettings
			paths, err := aas.SetFromFlags(cmd.Flags(), "")
			if err != nil {
				return err
			}
			js, err := api.Dial(ctx, config.JoinServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationActivationSettingRegistryClient(js).Set(ctx, &ttnpb.SetApplicationActivationSettingsRequest{
				ApplicationIds: appID,
				Settings:       &aas,
				FieldMask:      ttnpb.FieldMask(paths...),
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationActivationSettingsDeleteCommand = &cobra.Command{
		Use:     "delete [application-id]",
		Aliases: []string{"del", "remove", "rm"},
		Short:   "Delete application activation settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID.New()
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewApplicationActivationSettingRegistryClient(as).Delete(ctx, &ttnpb.DeleteApplicationActivationSettingsRequest{
				ApplicationIds: appID,
			})
			return err
		},
	}
)

func init() {
	ttnpb.AddSelectFlagsForApplicationActivationSettings(selectApplicationActivationSettingsFlags, "", false)
	applicationActivationSettingsGetCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationActivationSettingsGetCommand.Flags().AddFlagSet(selectApplicationActivationSettingsFlags)
	applicationActivationSettingsGetCommand.Flags().AddFlagSet(selectAllApplicationActivationSettingsFlags)
	applicationActivationSettingsCommand.AddCommand(applicationActivationSettingsGetCommand)
	applicationActivationSettingsSetCommand.Flags().AddFlagSet(applicationIDFlags())
	ttnpb.AddSetFlagsForApplicationActivationSettings(applicationActivationSettingsSetCommand.Flags(), "", false)
	applicationActivationSettingsCommand.AddCommand(applicationActivationSettingsSetCommand)
	applicationActivationSettingsDeleteCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationActivationSettingsCommand.AddCommand(applicationActivationSettingsDeleteCommand)
	applicationsCommand.AddCommand(applicationActivationSettingsCommand)
}
