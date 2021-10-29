// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"time"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/cleanup"
	js "go.thethings.network/lorawan-stack/v3/pkg/joinserver"
	jsredis "go.thethings.network/lorawan-stack/v3/pkg/joinserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// NewJSDeviceRegistryCleaner returns a new instance of Join Server RegistryCleaner with a local set
// of devices.
func NewJSRegistryCleaner(ctx context.Context, config *redis.Config) (*js.RegistryCleaner, error) {
	deviceRegistry := &jsredis.DeviceRegistry{
		Redis:   redis.New(config.WithNamespace("js", "devices")),
		LockTTL: defaultLockTTL,
	}
	if err := deviceRegistry.Init(ctx); err != nil {
		return nil, shared.ErrInitializeApplicationServer.WithCause(err)
	}
	applicationActivationSettingRegistry := &jsredis.ApplicationActivationSettingRegistry{
		Redis:   redis.New(config.WithNamespace("js", "application-activation-settings")),
		LockTTL: defaultLockTTL,
	}
	if err := applicationActivationSettingRegistry.Init(ctx); err != nil {
		return nil, shared.ErrInitializeJoinServer.WithCause(err)
	}
	cleaner := &js.RegistryCleaner{
		DevRegistry:   deviceRegistry,
		AppAsRegistry: applicationActivationSettingRegistry,
	}
	err := cleaner.RangeToLocalSet(ctx)
	if err != nil {
		return nil, err
	}
	return cleaner, nil
}

var (
	jsDBCommand = &cobra.Command{
		Use:   "js-db",
		Short: "Manage Join Server database",
	}
	jsDBCleanupCommand = &cobra.Command{
		Use:   "cleanup",
		Short: "Clean stale Join Server application and device data",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("Initializing device and application activation settings registries")
			jsCleaner, err := NewJSRegistryCleaner(ctx, &config.Redis)
			if err != nil {
				return err
			}
			// Define retry delay for obtaining cluster peer connection.
			retryDelay := time.Duration(500) * time.Millisecond
			// Create cluster and grpc connection with identity server.
			conn, cl, err := NewClusterComponentConnection(ctx, *config, retryDelay, 5, ttnpb.ClusterRole_ENTITY_REGISTRY)
			if err != nil {
				return err
			}
			defer func() {
				logger.Debug("Leaving cluster...")
				if err := cl.Leave(); err != nil {
					logger.WithError(err).Error("Could not leave cluster")
					return
				}
				logger.Debug("Left cluster")
			}()
			paginationDelay, err := cmd.Flags().GetDuration("pagination-delay")
			if err != nil {
				return err
			}
			client := ttnpb.NewApplicationRegistryClient(conn)
			applicationList, err := FetchIdentityServerApplications(ctx, client, cl.Auth(), paginationDelay)
			if err != nil {
				return err
			}
			applicationIdentityServerSet := make(map[string]struct{})
			for _, app := range applicationList {
				applicationIdentityServerSet[unique.ID(ctx, app.GetIds())] = struct{}{}
			}
			devClient := ttnpb.NewEndDeviceRegistryClient(conn)
			endDeviceList, err := FetchIdentityServerEndDevices(ctx, devClient, cl.Auth(), paginationDelay)
			if err != nil {
				return err
			}
			deviceIdentityServerSet := make(map[string]struct{})
			for _, dev := range endDeviceList {
				deviceIdentityServerSet[unique.ID(ctx, dev.EndDeviceIdentifiers)] = struct{}{}
			}
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}
			// If dry run flag set, print the app data to be deleted.
			if dryRun {
				deviceSet := cleanup.ComputeSetComplement(deviceIdentityServerSet, jsCleaner.LocalDeviceSet)
				logger.Info("Deleting device registry data for devices: ", setToArray(deviceSet))
				applicationSet := cleanup.ComputeSetComplement(applicationIdentityServerSet, jsCleaner.LocalApplicationSet)
				logger.Info("Deleting application activation settings registry data for applications: ", setToArray(applicationSet))
				logger.Warn("Dry run finished. No data deleted.")
				return nil
			}
			logger.Info("Cleaning join server registries")
			err = jsCleaner.CleanData(ctx, deviceIdentityServerSet, applicationIdentityServerSet)
			if err != nil {
				return err
			}
			return nil
		},
	}
)

func init() {
	Root.AddCommand(jsDBCommand)
	jsDBCleanupCommand.Flags().Bool("dry-run", false, "Dry run")
	jsDBCleanupCommand.Flags().Duration("pagination-delay", 100, "Delay between batch requests")
	jsDBCommand.AddCommand(jsDBCleanupCommand)
}
