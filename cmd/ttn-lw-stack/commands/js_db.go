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
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/cleanup"
	js "go.thethings.network/lorawan-stack/v3/pkg/joinserver"
	jsredis "go.thethings.network/lorawan-stack/v3/pkg/joinserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
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
			devicesCl := NewJoinServerDeviceRegistryRedis(config)
			keysCl := NewJoinServerSessionKeyRegistryRedis(config)

			schemaVersion, err := getSchemaVersion(keysCl)
			if err != nil {
				return err
			}
			force, _ := cmd.Flags().GetBool("force")
			if !force {
				if schemaVersion >= jsredis.SchemaVersion {
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

			sessionKeyIDRegexp := regexp.MustCompile(keysCl.Key("id", euiRegexpStr, euiRegexpStr, sessionKeyIDStr+"$"))

			// Delete session keys that are not belonging to an existing device (orphan session key) and session keys that are
			// older than the configured limit.
			if (schemaVersion < 1 || force) && config.JS.SessionKeyLimit > 0 {
				sessionKeyIDs := make(map[string][][]byte)
				sessionKeyIDCount := 0

				// Load all session keys in memory and put them in a set for each end device.
				err = ttnredis.RangeRedisKeys(ctx, keysCl, keysCl.Key("*"), ttnredis.DefaultRangeCount, func(k string) (bool, error) {
					logger := logger.WithField("key", k)
					if match := sessionKeyIDRegexp.FindStringSubmatch(k); len(match) > 0 {
						sid, err := base64.RawStdEncoding.DecodeString(match[3])
						if err != nil {
							logger.WithError(err).Error("Failed to parse base64 session key ID")
							return true, nil
						}
						k := fmt.Sprintf("%s:%s", match[1], match[2])
						sessionKeyIDs[k] = append(sessionKeyIDs[k], sid)
						sessionKeyIDCount++
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
				logger.WithFields(log.Fields(
					"end_device_count", len(sessionKeyIDs),
					"session_key_count", sessionKeyIDCount,
				)).Debug("Found session keys")

				// Check whether the end device exists. If not, delete the session keys. Otherwise, delete the old session keys.
				for k, sids := range sessionKeyIDs {
					parts := strings.SplitN(k, ":", 2)
					joinEUI, devEUI := parts[0], parts[1]
					logger := logger.WithFields(log.Fields(
						"join_eui", joinEUI,
						"dev_eui", devEUI,
					))
					exists, err := devicesCl.Exists(ctx, devicesCl.Key("eui", joinEUI, devEUI)).Result()
					if err != nil {
						return err
					}
					if exists == 0 {
						_, err := keysCl.Pipelined(ctx, func(p redis.Pipeliner) error {
							for _, sid := range sids {
								p.Del(ctx, keysCl.Key("id", joinEUI, devEUI, base64.RawStdEncoding.EncodeToString(sid)))
							}
							return nil
						})
						if err != nil {
							logger.WithError(err).Error("Failed to delete orphan session key(s)")
							return err
						}
						logger.WithField("count", len(sids)).Debug("Deleted orphan session keys")
					} else {
						sort.Slice(sids, func(i, j int) bool { return bytes.Compare(sids[i], sids[j]) < 0 })
						if d := len(sids) - config.JS.SessionKeyLimit; d > 0 {
							_, err := keysCl.Pipelined(ctx, func(p redis.Pipeliner) error {
								for _, sid := range sids[:d] {
									p.Del(ctx, keysCl.Key("id", joinEUI, devEUI, base64.RawStdEncoding.EncodeToString(sid)))
								}
								return nil
							})
							if err != nil {
								logger.WithError(err).Error("Failed to delete old session key(s)")
								continue
							}
							logger.WithField("count", len(sids[:d])).Debug("Deleted old session keys")
							sids = sids[d:]
						}
						sidVals := make([]any, len(sids))
						for i := range sids {
							sidVals[i] = base64.RawStdEncoding.EncodeToString(sids[i])
						}
						sidsKey := keysCl.Key("ids", joinEUI, devEUI)
						if err := keysCl.Del(ctx, sidsKey).Err(); err != nil {
							logger.WithError(err).Error("Failed to delete existing recent session key IDs")
							return err
						}
						if err := keysCl.RPush(ctx, sidsKey, sidVals...).Err(); err != nil {
							logger.WithError(err).Error("Failed to store recent session key IDs")
							return err
						}
						logger.WithField("count", len(sidVals)).Debug("Stored recent session key IDs")
					}
				}
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
