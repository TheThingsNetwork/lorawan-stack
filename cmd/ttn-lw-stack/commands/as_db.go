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
	"regexp"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web"
	asredis "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/cleanup"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

var (
	asDBCommand = &cobra.Command{
		Use:   "as-db",
		Short: "Manage Application Server database",
	}
	asDBMigrateCommand = &cobra.Command{
		Use:   "migrate",
		Short: "Migrate Application Server data",
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.Redis.IsZero() {
				panic("Only Redis is supported by this command")
			}

			logger.Info("Connecting to Redis database...")
			cl := ttnredis.New(config.Redis.WithNamespace("as"))

			if force, _ := cmd.Flags().GetBool("force"); !force {
				schemaVersion, err := getSchemaVersion(cl)
				if err != nil {
					return err
				}
				if schemaVersion >= asredis.SchemaVersion {
					logger.Info("Database schema version is already in latest version")
					return nil
				}
			}

			var migrated uint64
			defer func() { logger.Debugf("%d keys migrated", migrated) }()

			const (
				idRegexpStr        = `([a-z0-9](?:[-]?[a-z0-9]){2,}){1,36}?`
				uidRegexpStr       = idRegexpStr
				deviceUIDRegexpStr = idRegexpStr + `\.` + uidRegexpStr
			)

			deviceUIDRegexp := regexp.MustCompile(cl.Key("devices", "uid", deviceUIDRegexpStr+"$"))

			lockerID, err := ttnredis.GenerateLockerID()
			if err != nil {
				return err
			}
			if err := ttnredis.RangeRedisKeys(ctx, cl, cl.Key("*"), ttnredis.DefaultRangeCount, func(k string) (bool, error) {
				logger := logger.WithField("key", k)
				switch {
				case deviceUIDRegexp.MatchString(k):
					if err := ttnredis.LockedWatch(ctx, cl, k, lockerID, defaultLockTTL, func(tx *redis.Tx) error {
						dev := &ttnpb.EndDevice{}
						if err := ttnredis.GetProto(ctx, tx, k).ScanProto(dev); err != nil {
							logger.WithError(err).Error("Failed to get device proto")
							return err
						}
						var any bool
						for _, sess := range []*ttnpb.Session{dev.Session, dev.PendingSession} {
							if sess == nil || sess.StartedAt == nil {
								continue
							}
							sess.StartedAt = nil
							any = true
						}
						if any {
							_, err := ttnredis.SetProto(ctx, tx, k, dev, 0)
							if err != nil {
								return err
							}
							migrated++
						}
						return nil
					}); err != nil {
						logger.WithError(err).Error("Transaction failed")
					}
				}
				return true, nil
			}); err != nil {
				return err
			}

			return recordSchemaVersion(cl, asredis.SchemaVersion)
		},
	}
	asDBCleanupCommand = &cobra.Command{
		Use:   "cleanup",
		Short: "Clean stale Application Server application data",
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.Redis.IsZero() {
				panic("Only Redis is supported by this command")
			}
			// Initialize AS registry cleaners (together with their local app/dev sets).
			logger.Info("Initiating PubSub client")
			pubsubCleaner, err := NewPubSubCleaner(ctx, &config.Redis)
			if err != nil {
				return err
			}
			webhookCleaner := &web.RegistryCleaner{}
			if config.AS.Webhooks.Target != "" {
				logger.Info("Initiating webhook client")
				webhookCleaner, err = NewWebhookCleaner(ctx, &config.Redis)
				if err != nil {
					return err
				}
			}
			logger.Info("Initiating application packages registry")
			appPackagesCleaner, err := NewPackagesCleaner(ctx, &config.Redis)
			if err != nil {
				return err
			}
			logger.Info("Initiating device registry")
			deviceCleaner, err := NewASDeviceRegistryCleaner(ctx, &config.Redis)
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
				deviceIdentityServerSet[unique.ID(ctx, dev.Ids)] = struct{}{}
			}
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}
			// If dry run flag set, print the app data to be deleted.
			if dryRun {
				logger.Warn("Command is running in dry run mode")
				pubsubAppSet := cleanup.ComputeSetComplement(applicationIdentityServerSet, pubsubCleaner.LocalSet)
				logger.Info("Deleting pubsub registry data for applications: ", setToArray(pubsubAppSet))

				webhookAppSet := cleanup.ComputeSetComplement(applicationIdentityServerSet, webhookCleaner.LocalSet)
				logger.Info("Deleting webhook registry data for applications: ", setToArray(webhookAppSet))

				appPackagesAppSet := cleanup.ComputeSetComplement(applicationIdentityServerSet, appPackagesCleaner.LocalApplicationSet)
				logger.Info("Deleting application packages registry data for applications: ", setToArray(appPackagesAppSet))

				appPackagesDevSet := cleanup.ComputeSetComplement(deviceIdentityServerSet, appPackagesCleaner.LocalDeviceSet)
				logger.Info("Deleting application packages registry data for devices: ", setToArray(appPackagesDevSet))

				deviceSet := cleanup.ComputeSetComplement(deviceIdentityServerSet, deviceCleaner.LocalSet)
				logger.Info("Deleting device registry data for devices: ", setToArray(deviceSet))

				logger.Warn("Dry run finished. No data deleted.")
				return nil
			}
			// Cleanup data from AS registries.
			logger.Info("Cleaning PubSub registry")
			err = pubsubCleaner.CleanData(ctx, applicationIdentityServerSet)
			if err != nil {
				return err
			}
			logger.Info("Cleaning application packages registry")
			err = appPackagesCleaner.CleanData(ctx, deviceIdentityServerSet, applicationIdentityServerSet)
			if err != nil {
				return err
			}
			logger.Info("Cleaning device registry")
			err = deviceCleaner.CleanData(ctx, deviceIdentityServerSet)
			if err != nil {
				return err
			}
			if webhookCleaner.WebRegistry != nil {
				logger.Info("Cleaning webhook registry")
				err = webhookCleaner.CleanData(ctx, applicationIdentityServerSet)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
)

func init() {
	Root.AddCommand(asDBCommand)
	asDBMigrateCommand.Flags().Bool("force", false, "Force perform database migrations")
	asDBCommand.AddCommand(asDBMigrateCommand)
	asDBCleanupCommand.Flags().Bool("dry-run", false, "Dry run")
	asDBCleanupCommand.Flags().Duration("pagination-delay", 100, "Delay between batch requests")
	asDBCommand.AddCommand(asDBCleanupCommand)
}
