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
	"time"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/pkg/cleanup"
	nsredis "go.thethings.network/lorawan-stack/v3/pkg/networkserver/redis"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

var (
	nsDBCommand = &cobra.Command{
		Use:   "ns-db",
		Short: "Manage Network Server database",
	}
	nsDBPruneCommand = &cobra.Command{
		Use:   "prune",
		Short: "Remove unused Network Server data",
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.Redis.IsZero() {
				panic("Only Redis is supported by this command")
			}

			logger.Info("Connecting to Redis database...")
			cl := NewNetworkServerApplicationUplinkQueueRedis(config)
			var deleted uint64
			defer func() { logger.Debugf("%d processed stream entries deleted", deleted) }()
			return ttnredis.RangeRedisKeys(ctx, cl, nsredis.ApplicationUplinkQueueUIDGenericUplinkKey(cl, "*"), ttnredis.DefaultRangeCount, func(k string) (bool, error) {
				gs, err := cl.XInfoGroups(ctx, k).Result()
				if err != nil {
					logger.WithError(err).Errorf("Failed to query groups of stream %q", k)
					return true, nil
				}
				for _, g := range gs {
					if g.Name != "ns" {
						logger.Errorf("Unexpected consumer group with name %q found for stream %q", g.Name, k)
						continue
					}
					last := "-"
					for {
						msgs, err := cl.XRangeN(ctx, k, last, g.LastDeliveredID, 1).Result()
						if err != nil {
							logger.WithError(err).Errorf("Failed to XRANGE over stream %q", k)
							return true, nil
						}
						if len(msgs) == 0 {
							return true, nil
						}
						var ids []string
						for _, msg := range msgs {
							ids = append(ids, msg.ID)
							last = msg.ID
						}
						_, err = cl.XDel(ctx, k, ids...).Result()
						if err != nil {
							logger.WithError(err).Errorf("Failed to XDEL from stream %q, continue to next stream", k)
							return true, nil
						}
						deleted += uint64(len(ids))
					}
				}
				return true, nil
			})
		},
	}
	nsDBMigrateCommand = &cobra.Command{
		Use:   "migrate",
		Short: "Migrate Network Server data",
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.Redis.IsZero() {
				panic("Only Redis is supported by this command")
			}

			logger.Info("Connecting to Network Server database...")
			devicesClient := NewNetworkServerDeviceRegistryRedis(config)
			uplinkClient := NewNetworkServerApplicationUplinkQueueRedis(config)
			defer devicesClient.Close()
			defer uplinkClient.Close()

			if force, _ := cmd.Flags().GetBool("force"); !force {
				logger.Info("Checking devices namespace schema version...")
				devicesSchemaVersion, err := getSchemaVersion(devicesClient)
				if err != nil {
					return err
				}

				if devicesSchemaVersion >= nsredis.DeviceSchemaVersion {
					logger.Info("Devices namespace schema version is already in latest version")
				}

				if devicesSchemaVersion < nsredis.UnsupportedDeviceMigrationVersionBreakpoint {
					return fmt.Errorf( // nolint:stylecheck
						"You are currently running devices namespace schema version %d of the Network Server database. "+
							"This devices namespace schema cannot be auto-migrated to latest. "+
							"Use v3.24.0 of The Things Stack to upgrade the devices namespace version before upgrading to latest. "+
							"If you are not upgrading from a version preceding v3.11 of The Things Stack please use the --force flag.", // nolint:lll
						devicesSchemaVersion,
					)
				}

				logger.Info("Checking uplink namespace schema version...")
				uplinkSchemaVersion, err := getSchemaVersion(uplinkClient)
				if err != nil {
					return err
				}
				if uplinkSchemaVersion >= nsredis.DeviceSchemaVersion {
					logger.Info("Uplink namespace schema version is already in latest version")
					return nil
				}
			}

			var removed uint64
			defer func() { logger.Debugf("%d keys removed", removed) }()

			uidLastInvalidationKey := uplinkClient.Key("uid", "*", "last-invalidation")
			pipeliner := uplinkClient.Pipeline()
			err := ttnredis.RangeRedisKeys(ctx, uplinkClient, uidLastInvalidationKey, ttnredis.DefaultRangeCount,
				func(k string) (bool, error) {
					logger := logger.WithField("key", k)
					if err := pipeliner.Del(ctx, k).Err(); err != nil {
						logger.WithError(err).Error("Failed to delete key")
						return true, nil
					}
					removed++
					return true, nil
				})
			if err != nil {
				return err
			}
			if _, err := pipeliner.Exec(ctx); err != nil {
				return err
			}

			if err := recordSchemaVersion(uplinkClient, nsredis.UplinkSchemaVersion); err != nil {
				return err
			}
			return recordSchemaVersion(devicesClient, nsredis.DeviceSchemaVersion)
		},
	}
	nsDBCleanupCommand = &cobra.Command{
		Use:   "cleanup",
		Short: "Clean stale Network Server application data",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("Initiating device registry")
			deviceCleaner, err := NewNSDeviceRegistryCleaner(ctx, &config.Redis)
			if err != nil {
				return err
			}
			// Define retry delay for obtaining cluster peer connection.
			retryDelay := time.Duration(500) * time.Millisecond
			// Create cluster and grpc connection with identity server.
			conn, cl, err := NewClusterComponentConnection(ctx, config, retryDelay, 5, ttnpb.ClusterRole_ENTITY_REGISTRY)
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
			devClient := ttnpb.NewEndDeviceRegistryClient(conn)
			endDeviceList, err := FetchIdentityServerEndDevices(ctx, devClient, cl.Auth(), paginationDelay)
			if err != nil {
				return err
			}
			deviceIdentityServerSet := make(map[string]struct{})
			for _, dev := range endDeviceList {
				deviceIdentityServerSet[unique.ID(ctx, dev.Ids)] = struct{}{}
			}
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}
			// If dry run flag set, print the app data to be deleted.
			if dryRun {
				deviceSet := cleanup.ComputeSetComplement(deviceIdentityServerSet, deviceCleaner.LocalSet)
				logger.Info("Deleting device registry data for devices: ", setToArray(deviceSet))
				logger.Warn("Dry run finished. No data deleted.")
				return nil
			}
			logger.Info("Cleaning device registry")
			err = deviceCleaner.CleanData(ctx, deviceIdentityServerSet)
			if err != nil {
				return err
			}
			return nil
		},
	}
	nsDBPurgeCommand = &cobra.Command{
		Use:   "purge",
		Short: "Purge Network Server application data",
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.Redis.IsZero() {
				panic("Only Redis is supported by this command")
			}

			logger.Info("Connecting to Redis database...")
			cl := NewNetworkServerApplicationUplinkQueueRedis(config)
			defer cl.Close()

			var purged uint64

			genericUIDKeys := nsredis.ApplicationUplinkQueueUIDGenericUplinkKey(cl, "*")
			invalidationUIDKeys := ttnredis.Key(genericUIDKeys, "invalidation")
			joinAcceptUIDKeys := ttnredis.Key(genericUIDKeys, "join-accept")
			taskQueueKeys := ttnredis.Key(cl.Key("application"), "*")

			targets := []string{
				genericUIDKeys,
				invalidationUIDKeys,
				joinAcceptUIDKeys,
				taskQueueKeys,
			}

			pipeliner := cl.Pipeline()
			for _, target := range targets {
				err := ttnredis.RangeRedisKeys(ctx, cl, target, ttnredis.DefaultRangeCount,
					func(k string) (bool, error) {
						pipeliner.Del(ctx, k)
						purged++
						return true, nil
					})
				if err != nil {
					logger.WithError(err).Error("Failed to purge Network Server application data")
					return err
				}
			}
			if _, err := pipeliner.Exec(ctx); err != nil {
				logger.WithError(err).Error("Failed to purge Network Server application data")
				return err
			}

			logger.WithField("records_purged_count", purged).Info("Purged Network Server application data")
			return nil
		},
	}
)

func init() {
	Root.AddCommand(nsDBCommand)
	nsDBCommand.AddCommand(nsDBPruneCommand)
	nsDBMigrateCommand.Flags().Bool("force", false, "Force perform database migrations")
	nsDBCommand.AddCommand(nsDBMigrateCommand)
	nsDBCleanupCommand.Flags().Bool("dry-run", false, "Dry run")
	nsDBCleanupCommand.Flags().Duration("pagination-delay", 100, "Delay between batch requests")
	nsDBCommand.AddCommand(nsDBCleanupCommand)
	nsDBCommand.AddCommand(nsDBPurgeCommand)
}
