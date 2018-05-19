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
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/store"
)

const (
	// SeparatorByte is character used to separate the keys.
	SeparatorByte = ':'

	// Separator is SeparatorByte converted to a string.
	Separator = string(SeparatorByte)
)

// Store represents a Redis store.Interface implemntation
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
			Addr: conf.Address,
			DB:   conf.Database,
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

// Create stores generates an ULID and stores fields under a key associated with it.
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
			return errors.Errorf("A key %s already exists", key)
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

// Delete deletes the fields stored under the key associated with id.
func (s *Store) Delete(id store.PrimaryKey) (err error) {
	if id == nil {
		return store.ErrNilKey.New(nil)
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
			for i, curr := range idxCurrent {
				if curr != nil {
					p.SRem(s.key(s.config.IndexKeys[i], curr.(string)), id.String())
				}
			}
			p.Del(key)
			return nil
		})
		return err
	}, key)
}

// Update overwrites field values stored under PrimaryKey specified with values in diff and rebinds indexed keys present in diff.
func (s *Store) Update(id store.PrimaryKey, diff map[string][]byte) error {
	if id == nil {
		return store.ErrNilKey.New(nil)
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

// Find returns the fields stored under PrimaryKey specified.
func (s *Store) Find(id store.PrimaryKey) (map[string][]byte, error) {
	if id == nil {
		return nil, store.ErrNilKey.New(nil)
	}

	m, err := newStringBytesMapCmd(s.Redis.HGetAll(s.key(id.String()))).Result()
	if err != nil {
		return nil, err
	}
	return m, nil
}

// FindBy returns mapping of PrimaryKey -> fields, which match field values specified in filter. Filter represents an AND relation,
// meaning that only entries matching all the fields in filter should be returned.
func (s *Store) FindBy(filter map[string][]byte, count uint64, f func(store.PrimaryKey, map[string][]byte) bool) error {
	if len(filter) == 0 {
		return store.ErrEmptyFilter.New(nil)
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
		return errors.New("At least one index key must be specified")
	}

	return s.Redis.Watch(func(tx *redis.Tx) error {
		key := s.key(s.newID().String())

		n, err := tx.SInterStore(key, idxKeys...).Result()
		if err != nil {
			return err
		}
		defer tx.Del(key)
		if n == 0 {
			return nil
		}

		if count > math.MaxInt64 {
			count = math.MaxInt64
		}

		var ids []string
		var c uint64
		for {
			ids, c, err = tx.SScan(key, c, "", int64(count)).Result()
			if err != nil {
				return err
			}

			cmds := make(map[ulid.ULID]*stringBytesMapCmd, len(ids))
			_, err = tx.Pipelined(func(p redis.Pipeliner) error {
				for _, str := range ids {
					id, err := ulid.Parse(str)
					if err != nil {
						return errors.NewWithCausef(err, "Failed to parse %s as ULID, database inconsistent", str)
					}
					cmds[id] = newStringBytesMapCmd(p.HGetAll(s.key(str)))
				}
				return nil
			})
			if err != nil {
				return err
			}

		outer:
			for id, cmd := range cmds {
				m, err := cmd.Result()
				if err != nil {
					return err
				}
				if len(m) == 0 {
					continue
				}
				for _, k := range fieldFilter {
					if !bytes.Equal(m[k], filter[k]) {
						continue outer
					}
				}
				if !f(id, m) {
					return nil
				}
			}

			if c == 0 {
				break
			}
		}
		return nil
	}, idxKeys...)
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

// Put adds bs to set identified by id.
func (s *Store) Put(id store.PrimaryKey, bs ...[]byte) error {
	if id == nil {
		return store.ErrNilKey.New(nil)
	}
	if len(bs) == 0 {
		return nil
	}
	return s.put(id, bs...)
}

// CreateSet creates a new set, containing bs.
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

// FindSet returns set identified by id.
func (s *Store) FindSet(id store.PrimaryKey) (bs [][]byte, err error) {
	if id == nil {
		return nil, store.ErrNilKey.New(nil)
	}
	return bs, s.Redis.SMembers(s.key(id.String())).ScanSlice(&bs)
}

// Contains reports whether b is contained in set identified by id.
func (s *Store) Contains(id store.PrimaryKey, b []byte) (bool, error) {
	if id == nil {
		return false, store.ErrNilKey.New(nil)
	}
	return s.Redis.SIsMember(s.key(id.String()), b).Result()
}

// Remove removes bs from set identified by id.
func (s *Store) Remove(id store.PrimaryKey, bs ...[]byte) error {
	if id == nil {
		return store.ErrNilKey.New(nil)
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

// Append appends bs to list identified by id.
func (s *Store) Append(id store.PrimaryKey, bs ...[]byte) error {
	if id == nil {
		return store.ErrNilKey.New(nil)
	}
	if len(bs) == 0 {
		return nil
	}

	n, err := s.Redis.RPush(s.key(id.String()), bs).Result()
	if err != nil {
		return err
	}
	if n < int64(len(bs)) {
		return errors.Errorf("Expected to store %d values, stored %d", len(bs), n)
	}
	return nil
}

// CreateList creates a new list, containing bs.
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

// FindList returns list identified by id.
func (s *Store) FindList(id store.PrimaryKey) (bs [][]byte, err error) {
	if id == nil {
		return nil, store.ErrNilKey.New(nil)
	}
	return bs, s.Redis.LRange(s.key(id.String()), 0, -1).ScanSlice(&bs)
}

// Len returns length of the list identified by id.
func (s *Store) Len(id store.PrimaryKey) (int64, error) {
	if id == nil {
		return 0, store.ErrNilKey.New(nil)
	}
	return s.Redis.LLen(s.key(id.String())).Result()
}

// Pop returns the value stored at last index of list identified by id and removes it from the list.
func (s *Store) Pop(id store.PrimaryKey) (bs []byte, err error) {
	if id == nil {
		return nil, store.ErrNilKey.New(nil)
	}
	return s.Redis.LPop(s.key(id.String())).Bytes()
}
