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

package redis

import (
	"context"
	"fmt"
	"runtime/trace"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func rangeScanIteration(cmd *redis.ScanCmd, f func(...string) (bool, error)) (uint64, error) {
	ks, cursor, err := cmd.Result()
	if err != nil {
		return 0, err
	}
	ok, err := f(ks...)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, nil
	}
	return cursor, nil
}

func rangeScan(scan func(uint64) *redis.ScanCmd, f func(...string) (bool, error)) error {
	cursor, err := rangeScanIteration(scan(0), f)
	if err != nil {
		return err
	}
	for cursor > 0 {
		cursor, err = rangeScanIteration(scan(cursor), f)
		if err != nil {
			return err
		}
	}
	return nil
}

func rangeStrings(f func(string) (bool, error), ss ...string) (bool, error) {
	for _, s := range ss {
		ok, err := f(s)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func rangeStringsBindFunc(f func(string) (bool, error)) func(ss ...string) (bool, error) {
	return func(ss ...string) (bool, error) {
		return rangeStrings(f, ss...)
	}
}

// DefaultRangeCount is the default number of elements to be returned by a SCAN-family operation.
const DefaultRangeCount = 1024

func RangeRedisKeys(ctx context.Context, r redis.Cmdable, match string, count int64, f func(k string) (bool, error)) error {
	defer trace.StartRegion(ctx, "range keys").End()
	return rangeScan(func(cursor uint64) *redis.ScanCmd {
		return r.Scan(ctx, cursor, match, count)
	}, rangeStringsBindFunc(f))
}

func RangeRedisSet(ctx context.Context, r redis.Cmdable, scanKey, match string, count int64, f func(v string) (bool, error)) error {
	defer trace.StartRegion(ctx, "range set").End()
	return rangeScan(func(cursor uint64) *redis.ScanCmd {
		return r.SScan(ctx, scanKey, cursor, match, count)
	}, rangeStringsBindFunc(f))
}

func RangeRedisZSet(ctx context.Context, r redis.Cmdable, scanKey, match string, count int64, f func(k string, v float64) (bool, error)) error {
	defer trace.StartRegion(ctx, "range zset").End()
	return rangeScan(func(cursor uint64) *redis.ScanCmd {
		return r.ZScan(ctx, scanKey, cursor, match, count)
	}, func(ss ...string) (bool, error) {
		if n := len(ss); n%2 != 0 {
			panic(fmt.Sprintf("ZSCAN return value length is not even: %d", n))
		}
		for i := 0; i < len(ss); i += 2 {
			v, err := strconv.ParseFloat(ss[i+1], 64)
			if err != nil {
				panic(fmt.Sprintf("failed to parse element score as float64: %s", err))
			}
			ok, err := f(ss[i], v)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		}
		return true, nil
	})
}

func RangeRedisHMap(ctx context.Context, r redis.Cmdable, scanKey, match string, count int64, f func(k string, v string) (bool, error)) error {
	defer trace.StartRegion(ctx, "range hmap").End()
	return rangeScan(func(cursor uint64) *redis.ScanCmd {
		return r.HScan(ctx, scanKey, cursor, match, count)
	}, func(ss ...string) (bool, error) {
		if n := len(ss); n%2 != 0 {
			panic(fmt.Sprintf("HSCAN return value length is not even: %d", n))
		}
		for i := 0; i < len(ss); i += 2 {
			ok, err := f(ss[i], ss[i+1])
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		}
		return true, nil
	})
}
