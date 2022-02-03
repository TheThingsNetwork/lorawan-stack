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
	"bytes"
	"context"
	"encoding/base64"
	"regexp"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/cleanup"
	js "go.thethings.network/lorawan-stack/v3/pkg/joinserver"
	jsredis "go.thethings.network/lorawan-stack/v3/pkg/joinserver/redis"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// NewJSRegistryCleaner returns a new instance of Join Server RegistryCleaner with a local set
// of devices and applications.
func NewJSRegistryCleaner(ctx context.Context, config *ttnredis.Config) (*js.RegistryCleaner, error) {
	deviceRegistry := &jsredis.DeviceRegistry{
		Redis:   ttnredis.New(config.WithNamespace("js", "devices")),
		LockTTL: defaultLockTTL,
	}
	if err := deviceRegistry.Init(ctx); err != nil {
		return nil, shared.ErrInitializeApplicationServer.WithCause(err)
	}
	applicationActivationSettingRegistry := &jsredis.ApplicationActivationSettingRegistry{
		Redis:   ttnredis.New(config.WithNamespace("js", "application-activation-settings")),
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
	jsDBMigrateCommand = &cobra.Command{
		Use:   "migrate",
		Short: "Migrate Join Server data",
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.Redis.IsZero() {
				panic("Only Redis is supported by this command")
			}

			logger.Info("Connecting to Redis database...")
			devicesCl := NewJoinServerDeviceRegistryRedis(*config)
			keysCl := NewJoinServerSessionKeyRegistryRedis(*config)

			if force, _ := cmd.Flags().GetBool("force"); !force {
				needMigration, err := checkLatestSchemaVersion(keysCl, jsredis.SchemaVersion)
				if err != nil {
					return err
				}
				if !needMigration {
					logger.Info("Database schema version is already in latest version")
					return nil
				}
			}

			var migrated uint64
			defer func() { logger.Debugf("%d keys migrated", migrated) }()

			const (
				euiRegexpStr    = "([[:xdigit:]]{16})"
				sessionKeyIDStr = "([A-Za-z0-9+/]+)"
			)

			euiRegexp := regexp.MustCompile(devicesCl.Key("eui", euiRegexpStr, euiRegexpStr+"$"))
			sessionKeyIDRegexp := regexp.MustCompile(keysCl.Key("id", euiRegexpStr, euiRegexpStr, sessionKeyIDStr+"$"))

			err := ttnredis.RangeRedisKeys(ctx, devicesCl, devicesCl.Key("*"), ttnredis.DefaultRangeCount, func(k string) (bool, error) {
				logger := logger.WithField("key", k)
				if match := euiRegexp.FindSubmatch([]byte(k)); len(match) > 0 && config.JS.SessionKeyLimit > 0 {
					var joinEUI, devEUI types.EUI64
					if err := joinEUI.UnmarshalText(match[1]); err != nil {
						logger.WithError(err).Error("Failed to parse JoinEUI")
						return true, nil
					}
					if err := devEUI.UnmarshalText(match[2]); err != nil {
						logger.WithError(err).Error("Failed to parse DevEUI")
						return true, nil
					}
					var sids [][]byte
					err := ttnredis.RangeRedisKeys(ctx, keysCl, keysCl.Key("id", joinEUI.String(), devEUI.String(), "*"), ttnredis.DefaultRangeCount, func(k string) (bool, error) {
						match := sessionKeyIDRegexp.FindStringSubmatch(k)
						if len(match) == 0 {
							logger.WithField("key", k).Error("Failed to parse session key ID key")
							return true, nil
						}
						sid, err := base64.RawStdEncoding.DecodeString(match[3])
						if err != nil {
							logger.WithError(err).Error("Failed to parse base64 session key ID")
							return true, nil
						}
						sids = append(sids, sid)
						return true, nil
					})
					if err != nil {
						logger.WithError(err).Error("Failed to range session keys")
						return true, nil
					}
					logger.WithField("count", len(sids)).Debug("Retrieved session keys")
					sort.Slice(sids, func(i, j int) bool { return bytes.Compare(sids[i], sids[j]) < 0 })
					if d := len(sids) - config.JS.SessionKeyLimit; d > 0 {
						for _, sid := range sids[:d] {
							sidKey := keysCl.Key("id", joinEUI.String(), devEUI.String(), base64.RawStdEncoding.EncodeToString(sid))
							logger := logger.WithField("key", sidKey)
							if err := keysCl.Del(ctx, sidKey).Err(); err != nil {
								logger.WithError(err).Error("Failed to delete session key")
								continue
							}
							logger.Debug("Deleted old session key")
						}
						sids = sids[d:]
					}
					sidVals := make([]interface{}, len(sids))
					for i := range sids {
						sidVals[i] = base64.RawStdEncoding.EncodeToString(sids[i])
					}
					sidsKey := keysCl.Key("ids", joinEUI.String(), devEUI.String())
					if err := keysCl.Del(ctx, sidsKey).Err(); err != nil {
						logger.WithError(err).Error("Failed to delete existing recent session key IDs")
						return true, nil
					}
					if err := keysCl.RPush(ctx, sidsKey, sidVals...).Err(); err != nil {
						logger.WithError(err).Error("Failed to store recent session key IDs")
						return true, nil
					}
				} else {
					logger.Debug("Skip unmatched key")
					return true, nil
				}
				migrated++
				return true, nil
			})
			if err != nil {
				return err
			}

			return recordSchemaVersion(keysCl, jsredis.SchemaVersion)
		},
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
				deviceIdentityServerSet[unique.ID(ctx, dev.Ids)] = struct{}{}
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
	jsDBMigrateCommand.Flags().Bool("force", false, "Force perform database migrations")
	jsDBCommand.AddCommand(jsDBMigrateCommand)
	jsDBCleanupCommand.Flags().Bool("dry-run", false, "Dry run")
	jsDBCleanupCommand.Flags().Duration("pagination-delay", 100, "Delay between batch requests")
	jsDBCommand.AddCommand(jsDBCleanupCommand)
}
