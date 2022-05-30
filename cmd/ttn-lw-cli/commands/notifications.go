// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

	"github.com/TheThingsIndustries/protoc-gen-go-flags/flagsplugin"
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var notificationsCommand = &cobra.Command{
	Use:     "notifications",
	Aliases: []string{"notification"},
	Short:   "Manage notifications",
}

var notificationsListCommand = &cobra.Command{
	Use:   "list",
	Short: "List notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		req := &ttnpb.ListNotificationsRequest{
			Status: []ttnpb.NotificationStatus{
				ttnpb.NotificationStatus_NOTIFICATION_STATUS_UNSEEN,
			},
		}
		_, err := req.SetFromFlags(cmd.Flags(), "")
		if err != nil {
			return err
		}
		if len(args) > 0 && req.GetReceiverIds().GetUserId() == "" {
			if len(args) > 1 {
				logger.Warn("Multiple IDs found in arguments, considering only the first")
			}
			req.ReceiverIds = &ttnpb.UserIdentifiers{UserId: args[0]}
		}

		_, _, opt, getTotal := withPagination(cmd.Flags())

		is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
		if err != nil {
			return err
		}
		svc := ttnpb.NewNotificationServiceClient(is)

		res, err := svc.List(ctx, req, opt)
		if err != nil {
			return err
		}
		getTotal()

		if err = io.Write(os.Stdout, config.OutputFormat, res.Notifications); err != nil {
			return err
		}

		if markAsSeen, _ := cmd.Flags().GetBool("mark-as-seen"); markAsSeen {
			notificationIDs := make([]string, 0, len(res.Notifications))
			for _, notification := range res.Notifications {
				if notification.Status == ttnpb.NotificationStatus_NOTIFICATION_STATUS_UNSEEN {
					notificationIDs = append(notificationIDs, notification.Id)
				}
			}
			if len(notificationIDs) > 0 {
				_, err = svc.UpdateStatus(ctx, &ttnpb.UpdateNotificationStatusRequest{
					ReceiverIds: req.ReceiverIds,
					Ids:         notificationIDs,
					Status:      ttnpb.NotificationStatus_NOTIFICATION_STATUS_SEEN,
				})
				if err != nil {
					return err
				}
			}
		}

		return nil
	},
}

var notificationsSetStatusCommand = &cobra.Command{
	Use:     "set-status",
	Aliases: []string{"update-status"},
	Short:   "Set the status of notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		req := &ttnpb.UpdateNotificationStatusRequest{
			Status: ttnpb.NotificationStatus_NOTIFICATION_STATUS_SEEN,
		}
		_, err := req.SetFromFlags(cmd.Flags(), "")
		if err != nil {
			return err
		}
		if len(args) > 0 && req.GetReceiverIds().GetUserId() == "" {
			if len(args) > 1 {
				logger.Warn("Multiple IDs found in arguments, treating the first as user ID and the rest as notification IDs")
			}
			req.ReceiverIds = &ttnpb.UserIdentifiers{UserId: args[0]}
			args = args[1:]
		}
		if len(args) > 0 && len(req.GetIds()) == 0 {
			req.Ids = args
		}

		is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
		if err != nil {
			return err
		}
		svc := ttnpb.NewNotificationServiceClient(is)

		_, err = svc.UpdateStatus(ctx, req)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	ttnpb.AddSetFlagsForListNotificationsRequest(notificationsListCommand.Flags(), "", false)
	flagsplugin.AddAlias(notificationsListCommand.Flags(), "receiver-ids.user-id", "user-id", flagsplugin.WithHidden(false))
	notificationsListCommand.Flags().Lookup("status").DefValue = ttnpb.NotificationStatus_NOTIFICATION_STATUS_UNSEEN.String()
	notificationsListCommand.Flags().Bool("mark-as-seen", false, "Mark unseen notifications as seen")
	notificationsCommand.AddCommand(notificationsListCommand)
	ttnpb.AddSetFlagsForUpdateNotificationStatusRequest(notificationsSetStatusCommand.Flags(), "", false)
	flagsplugin.AddAlias(notificationsSetStatusCommand.Flags(), "receiver-ids.user-id", "user-id", flagsplugin.WithHidden(false))
	notificationsSetStatusCommand.Flags().Lookup("status").DefValue = ttnpb.NotificationStatus_NOTIFICATION_STATUS_SEEN.String()
	notificationsCommand.AddCommand(notificationsSetStatusCommand)
	Root.AddCommand(notificationsCommand)
}
