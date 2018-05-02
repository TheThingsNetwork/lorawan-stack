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
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/util/randutil"
	"github.com/oklog/ulid"
	redis "gopkg.in/redis.v5"
)

const (
	recursionLimit = 10

	// SeparatorByte is character used to separate the keys.
	SeparatorByte = ':'

	// Separator is SeparatorByte converted to a string.
	Separator = string(SeparatorByte)
)

var (
	_ store.ByteMapStore  = &Store{}
	_ store.ByteListStore = &Store{}
	_ store.ByteSetStore  = &Store{}
)

// Store represents a Redis store.Interface implemntation
type Store struct {
	Redis     *redis.Client
	config    *Config
	entropy   io.Reader
	indexKeys map[string]struct{}
}

// Config represents Redis configuration.
type Config struct {
	config.Redis
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
		entropy:   rand.New(randutil.NewLockedSource(rand.NewSource(time.Now().UnixNano()))),
		indexKeys: indexKeys,
	}
}

func (s *Store) key(str ...string) string {
	return s.config.Prefix + Separator + strings.Join(str, Separator)
}

func (s *Store) newID() fmt.Stringer {
	return ulid.MustNew(ulid.Now(), s.entropy)
}

// Create stores generates an ULID and stores fields under a key associated with it.
func (s *Store) Create(fields map[string][]byte) (store.PrimaryKey, error) {
	fieldsSet := make(map[string]string, len(fields))
	idxAdd := make([]string, 0, len(fields))
	for k, v := range fields {
		fieldsSet[k] = string(v)
		if _, ok := s.indexKeys[k]; ok {
			idxAdd = append(idxAdd, s.key(k, hex.EncodeToString(v)))
		}
	}

	id := s.newID()
	if len(fields) == 0 {
		return id, nil
	}

	idStr := id.String()
	key := s.key(idStr)

	// recursion levels
	var n int
	var create func() error
	create = func() error {
		err := s.Redis.Watch(func(tx *redis.Tx) error {
			ok, err := tx.Exists(key).Result()
			if err != nil {
				return err
			}
			if ok {
				return errors.Errorf("A key %s already exists", key)
			}
			_, err = tx.Pipelined(func(p *redis.Pipeline) error {
				for _, k := range idxAdd {
					p.SAdd(k, idStr)
				}
				if len(fieldsSet) != 0 {
					p.HMSet(key, fieldsSet)
				}
				return nil
			})
			return err
		}, key)
		if n != recursionLimit && err == redis.TxFailedErr {
			return create()
		}
		return err
	}
	return id, create()
}

// Delete deletes the fields stored under the key associated with id.
func (s *Store) Delete(id store.PrimaryKey) (err error) {
	if id == nil {
		return store.ErrNilKey.New(nil)
	}

	idStr := id.String()
	key := s.key(idStr)

	// recursion levels
	var n int
	var del func() error
	del = func() error {
		err = s.Redis.Watch(func(tx *redis.Tx) error {
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
			_, err = tx.Pipelined(func(p *redis.Pipeline) error {
				for i, curr := range idxCurrent {
					if curr != nil {
						p.SRem(s.key(s.config.IndexKeys[i], curr.(string)), idStr)
					}
				}
				p.Del(key)
				return nil
			})
			return err
		}, key)
		if n != recursionLimit && err == redis.TxFailedErr {
			return del()
		}
		return err
	}
	return del()
}

// Update overwrites field values stored under PrimaryKey specified with values in diff and rebinds indexed keys present in diff.
func (s *Store) Update(id store.PrimaryKey, diff map[string][]byte) (err error) {
	if id == nil {
		return store.ErrNilKey.New(nil)
	}
	if len(diff) == 0 {
		return nil
	}

	idxDel := make([]string, 0, len(diff))
	idxAdd := make([]string, 0, len(diff))
	fieldsDel := make([]string, 0, len(diff))
	fieldsSet := make(map[string]string, len(diff))

	for k, v := range diff {
		_, isIndex := s.indexKeys[k]
		if isIndex {
			idxDel = append(idxDel, k)
		}
		fieldsDel = append(fieldsDel, k)

		fieldsSet[k] = string(v)
		if isIndex {
			idxAdd = append(idxAdd, s.key(k, hex.EncodeToString(v)))
		}
	}

	idStr := id.String()
	key := s.key(idStr)

	// recursion levels
	var n int
	var update func() error
	update = func() error {
		err = s.Redis.Watch(func(tx *redis.Tx) error {
			var idxCurrent []interface{}
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

			_, err = tx.Pipelined(func(p *redis.Pipeline) error {
				for i, k := range idxDel {
					if curr := idxCurrent[i]; curr != nil {
						p.SRem(s.key(k, hex.EncodeToString([]byte(curr.(string)))), idStr)
					}
				}
				for _, k := range idxAdd {
					p.SAdd(k, idStr)
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
		if n != recursionLimit && err == redis.TxFailedErr {
			return update()
		}
		return err
	}
	return update()
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
func (s *Store) FindBy(filter map[string][]byte) (out map[store.PrimaryKey]map[string][]byte, err error) {
	if len(filter) == 0 {
		return nil, store.ErrEmptyFilter.New(nil)
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

	// recursion levels
	var n int
	var find func() error
	find = func() error {
		err := s.Redis.Watch(func(tx *redis.Tx) error {
			var ids []string
			if len(idxKeys) != 0 {
				ids, err = tx.SInter(idxKeys...).Result()
			} else {
				return errors.New("At least one index key must be specified")
			}
			if err != nil {
				return err
			}
			if len(ids) == 0 {
				return nil
			}

			cmds := make(map[ulid.ULID]*stringBytesMapCmd, len(ids))
			_, err = tx.Pipelined(func(p *redis.Pipeline) error {
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

			out = make(map[store.PrimaryKey]map[string][]byte, len(cmds))

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
				out[id] = m
			}
			if len(out) == 0 {
				out = nil
			}
			return nil
		}, idxKeys...)
		if n != recursionLimit && err == redis.TxFailedErr {
			return find()
		}
		return err
	}
	return out, find()
}

func (s *Store) put(id store.PrimaryKey, bs ...[]byte) error {
	k := s.key(id.String())
	_, err := s.Redis.Pipelined(func(p *redis.Pipeline) error {
		for _, b := range bs {
			p.SAdd(k, b)
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

	k := s.key(id.String())
	_, err := s.Redis.Pipelined(func(p *redis.Pipeline) error {
		for _, b := range bs {
			p.SRem(k, b)
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
