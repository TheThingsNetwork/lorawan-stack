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
	"fmt"
	stdio "io"
	"os"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func getStoredUp(cmd *cobra.Command, args []string, client ttnpb.ApplicationUpStorage_GetStoredApplicationUpClient, w stdio.Writer) error {
	streamOutput, _ := cmd.Flags().GetBool("stream-output")
	if !streamOutput {
		fmt.Fprintln(w, "[")
	}
	first := true
	for {
		up, err := client.Recv()
		switch err {
		case nil:
		case stdio.EOF:
			if !streamOutput {
				fmt.Fprintln(w, "]")
			}
			return nil
		default:
			return err
		}

		if !first && !streamOutput {
			fmt.Fprintln(w, ",")
		}
		if err := io.Write(w, config.OutputFormat, up); err != nil {
			return err
		}
		first = false
	}
}

var (
	endDevicesStorageCommand = &cobra.Command{
		Use:   "storage",
		Short: "Storage Integration",
	}
	endDeviceStorageGetCommand = &cobra.Command{
		Use:   "get [application-id] [device-id]",
		Short: "Retrieve stored upstream messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			req, err := getStoredUpRequest(cmd.Flags())
			if err != nil {
				return err
			}
			ids, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			req = req.WithEndDeviceIds(ids)
			client, err := ttnpb.NewApplicationUpStorageClient(as).GetStoredApplicationUp(ctx, req)
			if err != nil {
				return err
			}
			if err := getStoredUp(cmd, args, client, os.Stdout); err != nil {
				return err
			}
			return printContinuationToken(client, os.Stdout)
		},
	}
	endDeviceStorageCountCommand = &cobra.Command{
		Use:   "count [application-id] [device-id]",
		Short: "Count stored upstream messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			req, err := countStoredUpRequest(cmd.Flags())
			if err != nil {
				return err
			}
			ids, err := getEndDeviceID(cmd.Flags(), args, true)
			if err != nil {
				return err
			}
			req = req.WithEndDeviceIds(ids)
			resp, err := ttnpb.NewApplicationUpStorageClient(as).GetStoredApplicationUpCount(ctx, req)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, resp)
		},
	}

	applicationsStorageCommand = &cobra.Command{
		Use:   "storage",
		Short: "Storage Integration",
	}
	applicationsStorageGetCommand = &cobra.Command{
		Use:   "get [application-id]",
		Short: "Retrieve stored upstream messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			req, err := getStoredUpRequest(cmd.Flags())
			if err != nil {
				return err
			}
			ids := getApplicationID(cmd.Flags(), args)
			if ids == nil {
				return err
			}
			req = req.WithApplicationIds(ids)
			client, err := ttnpb.NewApplicationUpStorageClient(as).GetStoredApplicationUp(ctx, req)
			if err != nil {
				return err
			}

			if err := getStoredUp(cmd, args, client, os.Stdout); err != nil {
				return err
			}
			return printContinuationToken(client, os.Stdout)
		},
	}
	applicationsStorageCountCommand = &cobra.Command{
		Use:   "count [application-id]",
		Short: "Count stored upstream messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			as, err := api.Dial(ctx, config.ApplicationServerGRPCAddress)
			if err != nil {
				return err
			}
			req, err := countStoredUpRequest(cmd.Flags())
			if err != nil {
				return err
			}
			ids := getApplicationID(cmd.Flags(), args)
			if ids == nil {
				return err
			}
			req = req.WithApplicationIds(ids)
			resp, err := ttnpb.NewApplicationUpStorageClient(as).GetStoredApplicationUpCount(ctx, req)
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, resp)
		},
	}
)

func init() {
	endDeviceStorageGetCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDeviceStorageGetCommand.Flags().AddFlagSet(getStoredUpFlags())
	endDevicesStorageCommand.AddCommand(endDeviceStorageGetCommand)
	endDeviceStorageCountCommand.Flags().AddFlagSet(endDeviceIDFlags())
	endDeviceStorageCountCommand.Flags().AddFlagSet(countStoredUpFlags())
	endDevicesStorageCommand.AddCommand(endDeviceStorageCountCommand)
	endDevicesCommand.AddCommand(endDevicesStorageCommand)
	applicationsStorageGetCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsStorageGetCommand.Flags().AddFlagSet(getStoredUpFlags())
	applicationsStorageCommand.AddCommand(applicationsStorageGetCommand)
	applicationsStorageCountCommand.Flags().AddFlagSet(applicationIDFlags())
	applicationsStorageCountCommand.Flags().AddFlagSet(countStoredUpFlags())
	applicationsStorageCommand.AddCommand(applicationsStorageCountCommand)
	applicationsCommand.AddCommand(applicationsStorageCommand)
}
