// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package redis

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/randutil"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/oklog/ulid"
	redis "gopkg.in/redis.v5"
)

type Store struct {
	Redis     *redis.Client
	config    *Config
	entropy   io.Reader
	indexKeys map[string]struct{}
}

type Config struct {
	config.Redis
	Prefix    string
	IndexKeys []string
}

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
	return s.config.Prefix + ":" + strings.Join(str, ":")
}

func toString(v interface{}) (string, error) {
	switch t := v.(type) {
	case encoding.TextMarshaler:
		b, err := t.MarshalText()
		if err != nil {
			return "", err
		}
		return string(b), nil
	case encoding.BinaryMarshaler:
		b, err := t.MarshalBinary()
		if err != nil {
			return "", err
		}
		return string(b), nil
	case json.Marshaler:
		b, err := t.MarshalJSON()
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return fmt.Sprint(v), nil
}

func (s *Store) Create(fields map[string]interface{}) (store.PrimaryKey, error) {
	id := ulid.MustNew(ulid.Now(), s.entropy)
	idStr := id.String()

	_, err := s.Redis.Pipelined(func(p *redis.Pipeline) error {
		sfields := make(map[string]string, len(fields))
		for k, v := range fields {
			str, err := toString(v)
			if err != nil {
				return err
			}
			sfields[k] = str
			if _, ok := s.indexKeys[k]; ok {
				p.SAdd(s.key(k, str), idStr)
			}
		}
		p.HMSet(s.key(idStr), sfields)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return id, nil
}
func (s *Store) Delete(id store.PrimaryKey) error {
	indexKeys, err := s.Redis.HMGet(s.key(id.String()), s.config.IndexKeys...).Result()
	if err != nil {
		return err
	}

	_, err = s.Redis.Pipelined(func(p *redis.Pipeline) error {
		for i, v := range indexKeys {
			if v != nil {
				str, err := toString(v)
				if err != nil {
					return err
				}
				p.SRem(s.key(s.config.IndexKeys[i], str), id.String())
			}
		}
		p.Del(s.key(id.String()))
		return nil
	})
	return err
}

func (s *Store) Update(id store.PrimaryKey, diff map[string]interface{}) error {
	var (
		idStr = id.String()

		key = s.key(idStr)

		idxDel    = make([]string, 0, len(diff))
		idxSet    = make(map[string]string, len(diff))
		fieldsDel = make([]string, 0, len(diff))
		fieldsSet = make(map[string]string, len(diff))
	)
	for k, v := range diff {
		_, isIndex := s.indexKeys[k]
		if isIndex {
			idxDel = append(idxDel, k)
		}

		if v == nil {
			fieldsDel = append(fieldsDel, k)
			continue
		}

		str, err := toString(v)
		if err != nil {
			return err
		}
		fieldsSet[k] = str
		if isIndex {
			idxSet[k] = str
		}
	}

	idxCurrent, err := s.Redis.HMGet(s.key(id.String()), idxDel...).Result()
	if err != nil {
		return err
	}
	_, err = s.Redis.Pipelined(func(p *redis.Pipeline) error {
		p.HMSet(key, fieldsSet)
		p.HDel(key, fieldsDel...)
		for i, k := range idxDel {
			if idxCurrent[i] == nil {
				continue
			}
			p.SRem(s.key(k, idxCurrent[i].(string)), idStr)
		}
		for k, v := range idxSet {
			p.SAdd(s.key(k, v), idStr)
		}
		return nil
	})
	return err
}

type stringInterfaceMapCmd struct {
	*redis.StringStringMapCmd
}

func (c *stringInterfaceMapCmd) Result() (map[string]interface{}, error) {
	fields, err := c.StringStringMapCmd.Result()
	if err != nil {
		return nil, err
	}

	out := make(map[string]interface{}, len(fields))
	for k, v := range fields {
		out[k] = v
	}
	return out, nil
}

func newStringInterfaceMapCmd(c *redis.StringStringMapCmd) *stringInterfaceMapCmd {
	return &stringInterfaceMapCmd{c}
}

func (s *Store) Find(id store.PrimaryKey) (map[string]interface{}, error) {
	m, err := newStringInterfaceMapCmd(s.Redis.HGetAll(s.key(id.String()))).Result()
	if err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, store.ErrNotFound
	}
	return m, nil
}

func (s *Store) FindBy(filter map[string]interface{}) (map[store.PrimaryKey]map[string]interface{}, error) {
	keys := make([]string, 0, len(filter))
	for k, v := range filter {
		str, err := toString(v)
		if err != nil {
			return nil, err
		}
		keys = append(keys, s.key(k, str))
	}

	ids, err := s.Redis.SInter(keys...).Result()
	if err != nil {
		return nil, err
	}

	cmds := make(map[ulid.ULID]*stringInterfaceMapCmd, len(ids))
	if len(ids) > 0 {
		// Executing a pipeline with no commands throws an error
		_, err = s.Redis.Pipelined(func(p *redis.Pipeline) error {
			for _, str := range ids {
				id, err := ulid.Parse(str)
				if err != nil {
					return errors.New(fmt.Sprintf("pkg/store/redis: failed to parse %s as ULID, database inconsistent", str))
				}
				cmds[id] = newStringInterfaceMapCmd(p.HGetAll(s.key(str)))
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	out := make(map[store.PrimaryKey]map[string]interface{}, len(cmds))
	for id, cmd := range cmds {
		m, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		if len(m) == 0 {
			continue
		}
		out[id] = m
	}

	return out, nil
}
