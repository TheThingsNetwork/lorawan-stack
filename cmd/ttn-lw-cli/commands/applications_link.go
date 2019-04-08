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
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	selectApplicationLinkFlags = util.FieldMaskFlags(&ttnpb.ApplicationLink{})
	setApplicationLinkFlags    = util.FieldFlags(&ttnpb.ApplicationLink{})
)

var errNoApplicationLinkAPIKey = errors.DefineInvalidArgument("no_application_link_api_key", "no application link API key set")

var (
	applicationsLinkCommand = &cobra.Command{
		Use:   "link",
		Short: "Application link commands",
	}
	applicationsLinkGetCommand = &cobra.Command{
		Use:     "get [application-id]",
		Aliases: []string{"info"},
		Short:   "Get the properties of an application link",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectApplicationLinkFlags)
			if len(paths) == 0 {
				logger.Warn("No fields selected, will select everything")
				selectApplicationLinkFlags.VisitAll(func(flag *pflag.Flag) {
					paths = append(paths, strings.Replace(flag.Name, "-", "_", -1))
				})
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewAsClient(as).GetLink(ctx, &ttnpb.GetApplicationLinkRequest{
				ApplicationIdentifiers: *appID,
				FieldMask:              types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationsLinkSetCommand = &cobra.Command{
		Use:     "set [application-id]",
		Aliases: []string{"update"},
		Short:   "Set the properties of an application link",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}
			paths := util.UpdateFieldMask(cmd.Flags(), setApplicationLinkFlags)

			var link ttnpb.ApplicationLink
			if err := util.SetFields(&link, setApplicationLinkFlags); err != nil {
				return err
			}
			if link.APIKey == "" {
				return errNoApplicationLinkAPIKey
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewAsClient(as).SetLink(ctx, &ttnpb.SetApplicationLinkRequest{
				ApplicationIdentifiers: *appID,
				ApplicationLink:        link,
				FieldMask:              types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationsLinkDeleteCommand = &cobra.Command{
		Use:   "delete [application-id]",
		Short: "Delete an application link",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewAsClient(as).DeleteLink(ctx, appID)
			if err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	applicationsLinkGetCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsLinkGetCommand.Flags().AddFlagSet(selectApplicationLinkFlags)
	applicationsLinkCommand.AddCommand(applicationsLinkGetCommand)
	applicationsLinkSetCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsLinkSetCommand.Flags().AddFlagSet(setApplicationLinkFlags)
	applicationsLinkCommand.AddCommand(applicationsLinkSetCommand)
	applicationsLinkDeleteCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsLinkCommand.AddCommand(applicationsLinkDeleteCommand)
	applicationsCommand.AddCommand(applicationsLinkCommand)
}
