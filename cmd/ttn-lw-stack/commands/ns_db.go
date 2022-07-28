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
	"regexp"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
	"github.com/vmihailenco/msgpack/v5"
	"go.thethings.network/lorawan-stack/v3/pkg/cleanup"
	nsredis "go.thethings.network/lorawan-stack/v3/pkg/networkserver/redis"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
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

			logger.Info("Connecting to Redis database...")
			cl := NewNetworkServerDeviceRegistryRedis(config)

			if force, _ := cmd.Flags().GetBool("force"); !force {
				schemaVersion, err := getSchemaVersion(cl)
				if err != nil {
					return err
				}
				if schemaVersion >= nsredis.SchemaVersion {
					logger.Info("Database schema version is already in latest version")
					return nil
				}
			}

			var migrated uint64
			defer func() { logger.Debugf("%d keys migrated", migrated) }()

			type KeyEnvelope3_11Snapshot struct {
				Key          []byte `msgpack:"key"`
				KEKLabel     string `msgpack:"kek_label"`
				EncryptedKey []byte `msgpack:"encrypted_key"`
			}

			const (
				idRegexpStr      = `([a-z0-9](?:[-]?[a-z0-9]){2,}){1,36}?`
				uidRegexpStr     = idRegexpStr + `\.` + idRegexpStr
				euiRegexpStr     = "[[:xdigit:]]{16}"
				devAddrRegexpStr = "[[:xdigit:]]{8}"
			)

			uidRegexp := regexp.MustCompile(cl.Key("uid", uidRegexpStr+"$"))
			uidRegexp3_10_Fields := regexp.MustCompile(cl.Key("uid", uidRegexpStr, "fields$"))

			euiRegexp := regexp.MustCompile(cl.Key("eui", euiRegexpStr, euiRegexpStr+"$"))

			addrRegexpLegacy := regexp.MustCompile(cl.Key("addr", devAddrRegexpStr+"$"))
			addrRegexp3_10_16Bit := regexp.MustCompile(cl.Key("addr", devAddrRegexpStr, "16bit$"))
			addrRegexp3_10_32Bit := regexp.MustCompile(cl.Key("addr", devAddrRegexpStr, "32bit$"))
			addrRegexp3_10_Pending := regexp.MustCompile(cl.Key("addr", devAddrRegexpStr, "pending$"))
			addrRegexp3_11_Current := regexp.MustCompile(cl.Key("addr", devAddrRegexpStr, "current$"))
			addrRegexp3_11_CurrentFields := regexp.MustCompile(cl.Key("addr", devAddrRegexpStr, "current", "fields$"))
			addrRegexp3_11_PendingFields := regexp.MustCompile(cl.Key("addr", devAddrRegexpStr, "pending", "fields$"))
			err := ttnredis.RangeRedisKeys(ctx, cl, cl.Key("*"), ttnredis.DefaultRangeCount, func(k string) (bool, error) {
				logger := logger.WithField("key", k)
				switch {
				case uidRegexp3_10_Fields.MatchString(k):
					if err := cl.Del(ctx, k).Err(); err != nil {
						logger.WithError(err).Error("Failed to delete key")
						return true, nil
					}

				case addrRegexpLegacy.MatchString(k):
					var devAddr types.DevAddr
					if err := devAddr.UnmarshalText([]byte(k[len(k)-8:])); err != nil {
						logger.WithError(err).Error("Failed to parse DevAddr from legacy DevAddr key")
						return true, nil
					}
					currentKey := nsredis.CurrentAddrKey(k)
					currentFieldKey := nsredis.FieldKey(currentKey)
					pendingKey := nsredis.PendingAddrKey(k)
					pendingFieldKey := nsredis.FieldKey(pendingKey)
					pendingScore := float64(time.Now().UnixNano())
					if err := ttnredis.RangeRedisSet(ctx, cl, k, "*", ttnredis.DefaultRangeCount, func(uid string) (bool, error) {
						logger := logger.WithField("uid", uid)
						uk := nsredis.UIDKey(cl, uid)
						if err := cl.Watch(ctx, func(tx *redis.Tx) error {
							dev := &ttnpb.EndDevice{}
							if err := ttnredis.GetProto(ctx, tx, uk).ScanProto(dev); err != nil {
								logger.WithError(err).Error("Failed to get device proto")
								return err
							}
							p := tx.TxPipeline()
							defer p.Close()
							if dev.Session != nil && types.MustDevAddr(dev.Session.DevAddr).OrZero().Equal(devAddr) {
								if dev.MacState == nil {
									logger.Error("Device is missing MAC state, skip migrating current session")
								} else {
									b, err := nsredis.MarshalDeviceCurrentSession(dev)
									if err != nil {
										return err
									}
									p.ZAdd(ctx, currentKey, &redis.Z{
										Score:  float64(dev.Session.LastFCntUp & 0xffff),
										Member: uid,
									})
									p.HSet(ctx, currentFieldKey, uid, b)
								}
							}
							if dev.PendingSession != nil && types.MustDevAddr(dev.PendingSession.DevAddr).OrZero().Equal(devAddr) {
								if dev.PendingMacState == nil {
									logger.Error("Device is missing MAC state, skip migrating pending session")
								} else {
									b, err := nsredis.MarshalDevicePendingSession(dev)
									if err != nil {
										return err
									}
									p.ZAdd(ctx, pendingKey, &redis.Z{
										Score:  pendingScore,
										Member: uid,
									})
									p.HSet(ctx, pendingFieldKey, uid, b)
								}
							}
							p.SRem(ctx, k, uid)
							_, err := p.Exec(ctx)
							if err != nil {
								logger.WithError(err).Error("Pipeline failed")
								return err
							}
							return nil
						}, k, uk); err != nil {
							logger.WithError(err).Error("Transaction failed")
						}
						return true, nil
					}); err != nil {
						logger.WithError(err).Error("Failed to scan legacy DevAddr key")
						return true, nil
					}

				case addrRegexp3_10_16Bit.MatchString(k), addrRegexp3_10_32Bit.MatchString(k):
					currentKey := nsredis.CurrentAddrKey(k[:len(k)-6])
					fieldKey := nsredis.FieldKey(currentKey)
					if err := ttnredis.RangeRedisZSet(ctx, cl, k, "*", ttnredis.DefaultRangeCount, func(uid string, v float64) (bool, error) {
						logger := logger.WithField("uid", uid)
						uk := nsredis.UIDKey(cl, uid)
						if err := cl.Watch(ctx, func(tx *redis.Tx) error {
							dev := &ttnpb.EndDevice{}
							if err := ttnredis.GetProto(ctx, tx, uk).ScanProto(dev); err != nil {
								logger.WithError(err).Error("Failed to get device proto")
								return err
							}
							if dev.Session == nil || dev.MacState == nil {
								logger.Error("Device is missing session or MAC state, skip")
								return nil
							}
							b, err := nsredis.MarshalDeviceCurrentSession(dev)
							if err != nil {
								return err
							}
							_, err = tx.TxPipelined(ctx, func(p redis.Pipeliner) error {
								p.ZAdd(ctx, currentKey, &redis.Z{
									Score:  float64(uint32(v) & 0xffff),
									Member: uid,
								})
								p.HSet(ctx, fieldKey, uid, b)
								p.ZRem(ctx, k, uid)
								return nil
							})
							if err != nil {
								logger.WithError(err).Error("Pipeline failed")
								return err
							}
							return nil
						}, k, uk); err != nil {
							logger.WithError(err).Error("Transaction failed")
						}
						return true, nil
					}); err != nil {
						logger.WithError(err).Error("Failed to scan 3.10 current DevAddr key")
						return true, nil
					}

				case addrRegexp3_10_Pending.MatchString(k):
					typ, err := cl.Type(ctx, k).Result()
					if err != nil {
						logger.WithError(err).Error("Failed to determine type of value stored under key")
						return true, nil
					}
					if typ == "zset" {
						return true, nil
					}
					fieldKey := nsredis.FieldKey(k)
					tmpKey := cl.Key(k, "migrate")
					score := float64(time.Now().UnixNano())
					if err := cl.Watch(ctx, func(tx *redis.Tx) error {
						p := tx.TxPipeline()
						defer p.Close()
						if err := ttnredis.RangeRedisSet(ctx, tx, k, "*", ttnredis.DefaultRangeCount, func(uid string) (bool, error) {
							logger := logger.WithField("uid", uid)
							uk := nsredis.UIDKey(cl, uid)
							if err := tx.Watch(ctx, uk).Err(); err != nil {
								logger.WithField("key", uk).WithError(err).Error("Failed to watch UID key")
								return false, err
							}
							dev := &ttnpb.EndDevice{}
							if err := ttnredis.GetProto(ctx, tx, uk).ScanProto(dev); err != nil {
								logger.WithError(err).Error("Failed to get device proto")
								return false, err
							}
							if dev.PendingSession == nil || dev.PendingMacState == nil {
								logger.Error("Device is missing pending session or MAC state, skip")
								return true, nil
							}
							b, err := nsredis.MarshalDevicePendingSession(dev)
							if err != nil {
								return false, err
							}
							p.ZAdd(ctx, tmpKey, &redis.Z{
								Score:  score,
								Member: uid,
							})
							p.HSet(ctx, fieldKey, uid, b)
							p.SRem(ctx, k, uid)
							return true, nil
						}); err != nil {
							logger.WithError(err).Error("Failed to scan 3.10 pending DevAddr key")
							return err
						}
						p.RenameNX(ctx, tmpKey, k)
						_, err := p.Exec(ctx)
						if err != nil {
							logger.WithError(err).Error("Pipeline failed")
							return err
						}
						return nil
					}, k, tmpKey); err != nil {
						logger.WithError(err).Error("Transaction failed")
						return true, nil
					}

				case addrRegexp3_11_CurrentFields.MatchString(k):
					var migratedSubkeys uint64
					if err := ttnredis.RangeRedisHMap(ctx, cl, k, "*", 1, func(uid string, s string) (bool, error) {
						logger := logger.WithField("uid", uid)
						if err := cl.Watch(ctx, func(tx *redis.Tx) error {
							if err := msgpack.Unmarshal([]byte(s), &nsredis.UplinkMatchSession{}); err == nil {
								return nil
							}
							type BoolValue struct {
								Value bool `msgpack:"value"`
							}
							stored := &struct {
								FNwkSIntKey       KeyEnvelope3_11Snapshot
								ResetsFCnt        *BoolValue
								Supports32BitFCnt *BoolValue
								LoRaWANVersion    ttnpb.MACVersion
								LastFCnt          uint32
							}{}
							if err := msgpack.Unmarshal([]byte(s), stored); err != nil {
								logger.WithError(err).Error("Failed to unmarshal legacy current session fields")
								return err
							}
							var key *types.AES128Key
							if len(stored.FNwkSIntKey.Key) > 0 {
								key = &types.AES128Key{}
								copy(key[:], stored.FNwkSIntKey.Key)
							}
							var resetsFCnt *ttnpb.BoolValue
							if stored.ResetsFCnt != nil {
								resetsFCnt = &ttnpb.BoolValue{
									Value: stored.ResetsFCnt.Value,
								}
							}
							var supports32BitFCnt *ttnpb.BoolValue
							if stored.Supports32BitFCnt != nil {
								supports32BitFCnt = &ttnpb.BoolValue{
									Value: stored.Supports32BitFCnt.Value,
								}
							}
							b, err := msgpack.Marshal(&nsredis.UplinkMatchSession{
								FNwkSIntKey: &ttnpb.KeyEnvelope{
									Key:          key.Bytes(),
									KekLabel:     stored.FNwkSIntKey.KEKLabel,
									EncryptedKey: stored.FNwkSIntKey.EncryptedKey,
								},
								ResetsFCnt:        resetsFCnt,
								Supports32BitFCnt: supports32BitFCnt,
								LoRaWANVersion:    stored.LoRaWANVersion,
								LastFCnt:          stored.LastFCnt,
							})
							if err != nil {
								logger.WithError(err).Error("Failed to marshal current session fields")
								return nil
							}
							if err := tx.HSet(ctx, k, uid, string(b)).Err(); err != nil {
								logger.WithError(err).Error("Failed to set current session fields")
								return err
							}
							migratedSubkeys++
							return nil
						}, k); err != nil {
							logger.WithError(err).Error("Transaction failed")
						}
						return true, nil
					}); err != nil {
						logger.WithError(err).Error("Failed to scan current session field key")
						return true, nil
					}
					if migratedSubkeys == 0 {
						return true, nil
					}

				case addrRegexp3_11_PendingFields.MatchString(k):
					var migratedSubkeys uint64
					if err := ttnredis.RangeRedisHMap(ctx, cl, k, "*", 1, func(uid string, s string) (bool, error) {
						logger := logger.WithField("uid", uid)
						if err := cl.Watch(ctx, func(tx *redis.Tx) error {
							if err := msgpack.Unmarshal([]byte(s), &nsredis.UplinkMatchPendingSession{}); err == nil {
								return nil
							}
							stored := &struct {
								FNwkSIntKey    KeyEnvelope3_11Snapshot
								LoRaWANVersion ttnpb.MACVersion
							}{}
							if err := msgpack.Unmarshal([]byte(s), stored); err != nil {
								logger.WithError(err).Error("Failed to unmarshal legacy pending session fields")
								return err
							}
							var key *types.AES128Key
							if len(stored.FNwkSIntKey.Key) > 0 {
								key = &types.AES128Key{}
								copy(key[:], stored.FNwkSIntKey.Key)
							}
							b, err := msgpack.Marshal(&nsredis.UplinkMatchSession{
								FNwkSIntKey: &ttnpb.KeyEnvelope{
									Key:          key.Bytes(),
									KekLabel:     stored.FNwkSIntKey.KEKLabel,
									EncryptedKey: stored.FNwkSIntKey.EncryptedKey,
								},
								LoRaWANVersion: stored.LoRaWANVersion,
							})
							if err != nil {
								logger.WithError(err).Error("Failed to marshal pending session fields")
								return nil
							}
							if err := tx.HSet(ctx, k, uid, string(b)).Err(); err != nil {
								logger.WithError(err).Error("Failed to set pending session fields")
								return err
							}
							migratedSubkeys++
							return nil
						}, k); err != nil {
							logger.WithError(err).Error("Transaction failed")
						}
						return true, nil
					}); err != nil {
						logger.WithError(err).Error("Failed to scan pending session field key")
						return true, nil
					}
					if migratedSubkeys == 0 {
						return true, nil
					}

				case uidRegexp.MatchString(k),
					euiRegexp.MatchString(k),
					addrRegexp3_11_Current.MatchString(k):
					logger.Debug("Skip valid key")
					return true, nil

				default:
					d, err := cl.TTL(ctx, k).Result()
					if err != nil {
						logger.WithError(err).Error("Failed to determine TTL of unmatched key")
						return true, nil
					}
					if d < 0 {
						logger.Error("Skip unmatched key with no TTL")
						return true, nil
					}
					logger.Debug("Skip unmatched key with a TTL")
					return true, nil
				}
				logger.Debug("Migrated key to 3.11 format")
				migrated++
				return true, nil
			})
			if err != nil {
				return err
			}

			return recordSchemaVersion(cl, nsredis.SchemaVersion)
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
)

func init() {
	Root.AddCommand(nsDBCommand)
	nsDBCommand.AddCommand(nsDBPruneCommand)
	nsDBMigrateCommand.Flags().Bool("force", false, "Force perform database migrations")
	nsDBCommand.AddCommand(nsDBMigrateCommand)
	nsDBCleanupCommand.Flags().Bool("dry-run", false, "Dry run")
	nsDBCleanupCommand.Flags().Duration("pagination-delay", 100, "Delay between batch requests")
	nsDBCommand.AddCommand(nsDBCleanupCommand)
}
