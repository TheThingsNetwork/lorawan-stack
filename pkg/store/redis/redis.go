// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

// Package redis provides implementations of interfaces defined in store.
package redis

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/go-redis/redis"
	"github.com/oklog/ulid"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/store"
)

const (
	// SeparatorByte is character used to separate the keys.
	SeparatorByte = ':'

	// Separator is SeparatorByte converted to a string.
	Separator = string(SeparatorByte)
)

// Store represents a Redis store.ByteMapStore, store.ByteListStore and store.ByteListStore implementation.
type Store struct {
	Redis     *redis.Client
	config    *Config
	entropy   io.Reader
	namespace string
	indexKeys map[string]struct{}
}

// Config represents Redis configuration.
type Config struct {
	config.Redis
	Namespace []string
	IndexKeys []string
}

// New returns a new initialized Redis store.
func New(conf *Config) *Store {
	indexKeys := make(map[string]struct{}, len(conf.IndexKeys))
	for _, k := range conf.IndexKeys {
		indexKeys[k] = struct{}{}
	}
	return &Store{
		Redis: redis.NewClient(&redis.Options{
			Addr:     conf.Address,
			Password: conf.Password,
			DB:       conf.Database,
		}),
		config:    conf,
		entropy:   rand.Reader,
		namespace: strings.Join(append(conf.Redis.Namespace, conf.Namespace...), Separator),
		indexKeys: indexKeys,
	}
}

func (s *Store) key(str ...string) string {
	return s.namespace + Separator + strings.Join(str, Separator)
}

func (s *Store) newID() fmt.Stringer {
	return ulid.MustNew(ulid.Now(), s.entropy)
}

// Create implements store.ByteMapStore.
func (s *Store) Create(fields map[string][]byte) (store.PrimaryKey, error) {
	fieldsSet := make(map[string]interface{}, len(fields))
	idxAdd := make([]string, 0, len(fields))
	for k, v := range fields {
		fieldsSet[k] = v
		if _, ok := s.indexKeys[k]; ok {
			idxAdd = append(idxAdd, s.key(k, hex.EncodeToString(v)))
		}
	}

	id := s.newID()
	if len(fields) == 0 {
		return id, nil
	}

	key := s.key(id.String())
	return id, s.Redis.Watch(func(tx *redis.Tx) error {
		i, err := tx.Exists(key).Result()
		if err != nil {
			return err
		}
		if i != 0 {
			return store.ErrKeyAlreadyExists.WithAttributes("key", key)
		}
		_, err = tx.Pipelined(func(p redis.Pipeliner) error {
			for _, k := range idxAdd {
				p.SAdd(k, id.String())
			}
			if len(fieldsSet) != 0 {
				p.HMSet(key, fieldsSet)
			}
			return nil
		})
		return err
	}, key)
}

// Delete implements store.Deleter.
func (s *Store) Delete(id store.PrimaryKey) (err error) {
	if id == nil {
		return store.ErrNilKey
	}

	key := s.key(id.String())
	return s.Redis.Watch(func(tx *redis.Tx) error {
		var idxCurrent []interface{}
		if len(s.config.IndexKeys) != 0 {
			typ, err := tx.Type(key).Result()
			if err != nil {
				return err
			}
			if typ == "hash" {
				idxCurrent, err = tx.HMGet(key, s.config.IndexKeys...).Result()
				if err != nil {
					return err
				}
			}
		}
		_, err = tx.Pipelined(func(p redis.Pipeliner) error {
			for i, idx := range idxCurrent {
				if idx != nil {
					p.SRem(s.key(s.config.IndexKeys[i], hex.EncodeToString([]byte(idx.(string)))), id.String())
				}
			}
			p.Del(key)
			return nil
		})
		return err
	}, key)
}

// Update implements store.ByteMapStore.
func (s *Store) Update(id store.PrimaryKey, diff map[string][]byte) error {
	if id == nil {
		return store.ErrNilKey
	}
	if len(diff) == 0 {
		return nil
	}

	idxDel := make([]string, 0, len(diff))
	idxAdd := make([]string, 0, len(diff))
	fieldsDel := make([]string, 0, len(diff))
	fieldsSet := make(map[string]interface{}, len(diff))

	for k, v := range diff {
		_, isIndex := s.indexKeys[k]
		if isIndex {
			idxDel = append(idxDel, k)
		}
		fieldsDel = append(fieldsDel, k)

		fieldsSet[k] = v
		if isIndex {
			idxAdd = append(idxAdd, s.key(k, hex.EncodeToString(v)))
		}
	}

	key := s.key(id.String())
	return s.Redis.Watch(func(tx *redis.Tx) error {
		var idxCurrent []interface{}
		var err error
		if len(idxDel) != 0 {
			idxCurrent, err = tx.HMGet(key, idxDel...).Result()
			if err != nil {
				return err
			}
		}

		fieldsCurrent, err := tx.HKeys(key).Result()
		if err != nil {
			return err
		}

		curr := make(map[string]struct{}, len(fieldsCurrent))
		for _, k := range fieldsCurrent {
			curr[k] = struct{}{}
		}

		for _, dk := range fieldsDel {
			pre := dk + store.Separator
			for ck := range curr {
				if strings.HasPrefix(ck, pre) {
					fieldsDel = append(fieldsDel, ck)
					delete(curr, ck)
				}
			}
			if i := strings.LastIndexByte(dk, store.SeparatorByte); i != -1 {
				k := dk[:i]
				if _, ok := curr[k]; ok {
					fieldsDel = append(fieldsDel, k)
				}
			}
		}

		_, err = tx.Pipelined(func(p redis.Pipeliner) error {
			for i, k := range idxDel {
				if curr := idxCurrent[i]; curr != nil {
					p.SRem(s.key(k, hex.EncodeToString([]byte(curr.(string)))), id.String())
				}
			}
			for _, k := range idxAdd {
				p.SAdd(k, id.String())
			}
			if len(fieldsDel) != 0 {
				p.HDel(key, fieldsDel...)
			}
			if len(fieldsSet) != 0 {
				p.HMSet(key, fieldsSet)
			}
			return nil
		})
		return err
	}, key)
}

type stringBytesMapCmd struct {
	*redis.StringStringMapCmd
}

func (c *stringBytesMapCmd) Result() (map[string][]byte, error) {
	fields, err := c.StringStringMapCmd.Result()
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, nil
	}

	out := make(map[string][]byte, len(fields))
	for k, v := range fields {
		out[k] = []byte(v)
	}
	return out, nil
}

func newStringBytesMapCmd(c *redis.StringStringMapCmd) *stringBytesMapCmd {
	return &stringBytesMapCmd{c}
}

// Find implements store.ByteMapStore.
func (s *Store) Find(id store.PrimaryKey) (map[string][]byte, error) {
	if id == nil {
		return nil, store.ErrNilKey
	}

	m, err := newStringBytesMapCmd(s.Redis.HGetAll(s.key(id.String()))).Result()
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Range implements store.ByteMapStore.
func (s *Store) Range(filter map[string][]byte, orderBy string, count, offset uint64, f func(store.PrimaryKey, map[string][]byte) bool) (uint64, error) {
	if len(filter) == 0 {
		return 0, store.ErrEmptyFilter
	}

	idxKeys := make([]string, 0, len(filter))
	fieldFilter := make([]string, 0, len(filter))
	for k, v := range filter {
		if _, ok := s.indexKeys[k]; ok {
			idxKeys = append(idxKeys, s.key(k, hex.EncodeToString(v)))
		} else {
			fieldFilter = append(fieldFilter, k)
		}
	}
	if len(idxKeys) == 0 {
		return 0, errNoIndexKeys
	}

	if offset > math.MaxUint32 {
		return 0, errOffsetTooHigh
	}

	sort := &redis.Sort{
		By:     orderBy,
		Alpha:  true,
		Offset: int64(offset),
	}

	if len(fieldFilter) == 0 && count <= math.MaxInt64 {
		sort.Count = int64(count)
	}

	var cmds map[string]*stringBytesMapCmd
	var total int64
	if err := s.Redis.Watch(func(tx *redis.Tx) (err error) {
		key := s.key(s.newID().String())

		total, err = tx.SInterStore(key, idxKeys...).Result()
		if err != nil {
			return err
		}
		defer tx.Del(key)
		if total == 0 {
			return nil
		}

		ids, err := tx.Sort(key, sort).Result()
		if err != nil {
			return err
		}

		cmds = make(map[string]*stringBytesMapCmd, len(ids))
		_, err = tx.Pipelined(func(p redis.Pipeliner) error {
			for _, str := range ids {
				cmds[str] = newStringBytesMapCmd(p.HGetAll(s.key(str)))
			}
			return nil
		})
		return err
	}, idxKeys...); err != nil {
		return 0, err
	}

	// exec indicates whether f shall be executed or not.
	exec := true
outer:
	for str, cmd := range cmds {
		m, err := cmd.Result()
		if err != nil {
			return 0, err
		}

		if len(m) == 0 {
			total--
			continue
		}

		for _, k := range fieldFilter {
			if !bytes.Equal(m[k], filter[k]) {
				total--
				continue outer
			}
		}

		if !exec {
			continue
		}

		id, err := ulid.Parse(str)
		if err != nil {
			return 0, store.ErrInconsistentStore.WithCause(
				errParseULID.WithAttributes("ulid", str),
			)
		}

		if exec {
			exec = f(id, m)
		}
	}
	return uint64(total), nil
}

func (s *Store) put(id store.PrimaryKey, bs ...[]byte) error {
	key := s.key(id.String())
	_, err := s.Redis.Pipelined(func(p redis.Pipeliner) error {
		for _, b := range bs {
			p.SAdd(key, b)
		}
		return nil
	})
	return err
}

// Put implements store.ByteSetStore.
func (s *Store) Put(id store.PrimaryKey, bs ...[]byte) error {
	if id == nil {
		return store.ErrNilKey
	}
	if len(bs) == 0 {
		return nil
	}
	return s.put(id, bs...)
}

// CreateSet implements store.ByteSetStore.
func (s *Store) CreateSet(bs ...[]byte) (store.PrimaryKey, error) {
	id := s.newID()
	if len(bs) == 0 {
		return id, nil
	}

	if err := s.put(id, bs...); err != nil {
		return nil, err
	}
	return id, nil
}

// FindSet implements store.ByteSetStore.
func (s *Store) FindSet(id store.PrimaryKey) (bs [][]byte, err error) {
	if id == nil {
		return nil, store.ErrNilKey
	}
	return bs, s.Redis.SMembers(s.key(id.String())).ScanSlice(&bs)
}

// Contains implements store.ByteSetStore.
func (s *Store) Contains(id store.PrimaryKey, b []byte) (bool, error) {
	if id == nil {
		return false, store.ErrNilKey
	}
	return s.Redis.SIsMember(s.key(id.String()), b).Result()
}

// Remove implements store.ByteSetStore.
func (s *Store) Remove(id store.PrimaryKey, bs ...[]byte) error {
	if id == nil {
		return store.ErrNilKey
	}
	if len(bs) == 0 {
		return nil
	}

	_, err := s.Redis.Pipelined(func(p redis.Pipeliner) error {
		key := s.key(id.String())
		for _, b := range bs {
			p.SRem(key, b)
		}
		return nil
	})
	return err
}

// Append implements store.ByteListStore.
func (s *Store) Append(id store.PrimaryKey, bs ...[]byte) error {
	if id == nil {
		return store.ErrNilKey
	}
	if len(bs) == 0 {
		return nil
	}

	_, err := s.Redis.RPush(s.key(id.String()), bs).Result()
	return err
}

// CreateList implements store.ByteListStore.
func (s *Store) CreateList(bs ...[]byte) (store.PrimaryKey, error) {
	id := s.newID()
	if len(bs) == 0 {
		return id, nil
	}

	if err := s.Append(id, bs...); err != nil {
		return nil, err
	}
	return id, nil
}

// FindList implements store.ByteListStore.
func (s *Store) FindList(id store.PrimaryKey) (bs [][]byte, err error) {
	if id == nil {
		return nil, store.ErrNilKey
	}
	return bs, s.Redis.LRange(s.key(id.String()), 0, -1).ScanSlice(&bs)
}

// Len implements store.ByteListStore.
func (s *Store) Len(id store.PrimaryKey) (int64, error) {
	if id == nil {
		return 0, store.ErrNilKey
	}
	return s.Redis.LLen(s.key(id.String())).Result()
}

// Pop implements store.ByteListStore.
func (s *Store) Pop(id store.PrimaryKey) (bs []byte, err error) {
	if id == nil {
		return nil, store.ErrNilKey
	}
	return s.Redis.LPop(s.key(id.String())).Bytes()
}
