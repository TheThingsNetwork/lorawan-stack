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
	"github.com/spf13/cobra"
	nsredis "go.thethings.network/lorawan-stack/v3/pkg/networkserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/redis"
)

func rangeRedisKeysIteration(cl *redis.Client, cursor uint64, scanKey string, f func(k string) bool) (uint64, error) {
	ks, cursor, err := cl.Scan(cursor, scanKey, 1).Result()
	if err != nil {
		return 0, err
	}
	for _, k := range ks {
		if !f(k) {
			return 0, nil
		}
	}
	return cursor, nil
}

func rangeRedisKeys(cl *redis.Client, scanKey string, f func(k string) bool) error {
	cursor, err := rangeRedisKeysIteration(cl, 0, scanKey, f)
	if err != nil {
		return err
	}
	for cursor > 0 {
		cursor, err = rangeRedisKeysIteration(cl, cursor, scanKey, f)
		if err != nil {
			return err
		}
	}
	return nil
}

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
				panic("Only redis is supported by this command")
			}

			logger.Info("Connecting to Redis database...")
			cl := NewNetworkServerApplicationUplinkQueueRedis(*config)
			var deleted uint64
			defer func() { logger.Debugf("%d processed stream entries deleted", deleted) }()
			return rangeRedisKeys(cl, nsredis.ApplicationUplinkQueueUIDGenericUplinkKey(cl, "*"), func(k string) bool {
				gs, err := cl.XInfoGroups(k).Result()
				if err != nil {
					logger.WithError(err).Errorf("Failed to query groups of stream %q", k)
					return true
				}
				for _, g := range gs {
					if g.Name != "ns" {
						logger.Errorf("Unexpected consumer group with name %q found for stream %q", g.Name, k)
						continue
					}
					last := "-"
					for {
						msgs, err := cl.XRangeN(k, last, g.LastDeliveredID, 1).Result()
						if err != nil {
							logger.WithError(err).Errorf("Failed to XRANGE over stream %q", k)
							return true
						}
						if len(msgs) == 0 {
							return true
						}
						var ids []string
						for _, msg := range msgs {
							ids = append(ids, msg.ID)
							last = msg.ID
						}
						_, err = cl.XDel(k, ids...).Result()
						if err != nil {
							logger.WithError(err).Errorf("Failed to XDEL from stream %q, continue to next stream", k)
							return true
						}
						deleted += uint64(len(ids))
					}
				}
				return true
			})
		},
	}
)

func init() {
	Root.AddCommand(nsDBCommand)
	nsDBCommand.AddCommand(nsDBPruneCommand)
}
