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
	selectApplicationWebhookFlags = util.FieldMaskFlags(&ttnpb.ApplicationWebhook{})
	setApplicationWebhookFlags    = util.FieldFlags(&ttnpb.ApplicationWebhook{})
)

func applicationWebhookIDFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("application-id", "", "")
	flagSet.String("webhook-id", "", "")
	return flagSet
}

var errNoWebhookID = errors.DefineInvalidArgument("no_webhook_id", "no webhook ID set")

func getApplicationWebhookID(flagSet *pflag.FlagSet, args []string) (*ttnpb.ApplicationWebhookIdentifiers, error) {
	applicationID, _ := flagSet.GetString("application-id")
	webhookID, _ := flagSet.GetString("webhook-id")
	switch len(args) {
	case 0:
	case 1:
		logger.Warn("Only single ID found in arguments, not considering arguments")
	case 2:
		applicationID = args[0]
		webhookID = args[1]
	default:
		logger.Warn("Multiple IDs found in arguments, considering the first")
		applicationID = args[0]
		webhookID = args[1]
	}
	if applicationID == "" {
		return nil, errNoApplicationID
	}
	if webhookID == "" {
		return nil, errNoWebhookID
	}
	return &ttnpb.ApplicationWebhookIdentifiers{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: applicationID},
		WebhookID:              webhookID,
	}, nil
}

func headersFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.StringSlice("headers", nil, "key=value")
	return flagSet
}

var (
	applicationsWebhooksCommand = &cobra.Command{
		Use:     "webhooks",
		Aliases: []string{"webhook"},
		Short:   "Application webhooks commands",
	}
	applicationsWebhooksGetFormatsCommand = &cobra.Command{
		Use:     "get-formats",
		Aliases: []string{"formats"},
		Short:   "Get the available formats for application webhooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationWebhookRegistryClient(as).GetFormats(ctx, ttnpb.Empty)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationsWebhooksGetCommand = &cobra.Command{
		Use:     "get [application-id] [webhook-id]",
		Aliases: []string{"info"},
		Short:   "Get the properties of an application webhook",
		RunE: func(cmd *cobra.Command, args []string) error {
			webhookID, err := getApplicationWebhookID(cmd.Flags(), args)
			if err != nil {
				return err
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectApplicationWebhookFlags)
			if len(paths) == 0 {
				logger.Warn("No fields selected, will select everything")
				selectApplicationWebhookFlags.VisitAll(func(flag *pflag.Flag) {
					paths = append(paths, strings.Replace(flag.Name, "-", "_", -1))
				})
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationWebhookRegistryClient(as).Get(ctx, &ttnpb.GetApplicationWebhookRequest{
				ApplicationWebhookIdentifiers: *webhookID,
				FieldMask:                     types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationsWebhooksListCommand = &cobra.Command{
		Use:     "list [application-id]",
		Aliases: []string{"ls"},
		Short:   "List application webhooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			appID := getApplicationID(cmd.Flags(), args)
			if appID == nil {
				return errNoApplicationID
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectApplicationWebhookFlags)
			if len(paths) == 0 {
				logger.Warn("No fields selected, will select everything")
				selectApplicationWebhookFlags.VisitAll(func(flag *pflag.Flag) {
					paths = append(paths, strings.Replace(flag.Name, "-", "_", -1))
				})
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationWebhookRegistryClient(as).List(ctx, &ttnpb.ListApplicationWebhooksRequest{
				ApplicationIdentifiers: *appID,
				FieldMask:              types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationsWebhooksSetCommand = &cobra.Command{
		Use:     "set [application-id] [webhook-id]",
		Aliases: []string{"update"},
		Short:   "Set the properties of an application webhook",
		RunE: func(cmd *cobra.Command, args []string) error {
			webhookID, err := getApplicationWebhookID(cmd.Flags(), args)
			if err != nil {
				return err
			}
			paths := util.UpdateFieldMask(cmd.Flags(), setApplicationWebhookFlags, headersFlags())

			var webhook ttnpb.ApplicationWebhook
			if err = util.SetFields(&webhook, setApplicationWebhookFlags); err != nil {
				return err
			}
			headers, _ := cmd.Flags().GetStringSlice("headers")
			webhook.Headers = mergeKV(webhook.Headers, headers)
			webhook.ApplicationWebhookIdentifiers = *webhookID

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttnpb.NewApplicationWebhookRegistryClient(as).Set(ctx, &ttnpb.SetApplicationWebhookRequest{
				ApplicationWebhook: webhook,
				FieldMask:          types.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	applicationsWebhooksDeleteCommand = &cobra.Command{
		Use:   "delete [application-id] [webhook-id]",
		Short: "Delete an application webhook",
		RunE: func(cmd *cobra.Command, args []string) error {
			webhookID, err := getApplicationWebhookID(cmd.Flags(), args)
			if err != nil {
				return err
			}

			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttnpb.NewApplicationWebhookRegistryClient(as).Delete(ctx, webhookID)
			if err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	applicationsWebhooksCommand.AddCommand(applicationsWebhooksGetFormatsCommand)
	applicationsWebhooksGetCommand.Flags().AddFlagSet(applicationWebhookIDFlags())
	applicationsWebhooksGetCommand.Flags().AddFlagSet(selectApplicationWebhookFlags)
	applicationsWebhooksCommand.AddCommand(applicationsWebhooksGetCommand)
	applicationsWebhooksListCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsWebhooksListCommand.Flags().AddFlagSet(selectApplicationWebhookFlags)
	applicationsWebhooksCommand.AddCommand(applicationsWebhooksListCommand)
	applicationsWebhooksSetCommand.Flags().AddFlagSet(applicationWebhookIDFlags())
	applicationsWebhooksSetCommand.Flags().AddFlagSet(setApplicationWebhookFlags)
	applicationsWebhooksSetCommand.Flags().AddFlagSet(headersFlags())
	applicationsWebhooksCommand.AddCommand(applicationsWebhooksSetCommand)
	applicationsWebhooksDeleteCommand.Flags().AddFlagSet(applicationWebhookIDFlags())
	applicationsWebhooksCommand.AddCommand(applicationsWebhooksDeleteCommand)
	applicationsCommand.AddCommand(applicationsWebhooksCommand)
}
