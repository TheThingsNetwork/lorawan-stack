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

// Package ttnredis provides a general Redis client and component registry implementations.
package redis

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/gogo/protobuf/proto"
	"go.thethings.network/lorawan-stack/pkg/config"
)

const (
	// separator is character used to separate the keys.
	separator = ':'
)

var (
	encoding = base64.RawStdEncoding
)

// MarshalProto marshals pb into printable string.
func MarshalProto(pb proto.Message) (string, error) {
	b, err := proto.Marshal(pb)
	if err != nil {
		return "", err
	}
	return encoding.EncodeToString(b), nil
}

// MarshalProto unmarshals string returned from MarshalProto into pb.
func UnmarshalProto(s string, pb proto.Message) error {
	b, err := encoding.DecodeString(s)
	if err != nil {
		return err
	}
	return proto.Unmarshal(b, pb)
}

// Key constructs the full key for entity identified by ks by joining ks using the default separator.
func Key(ks ...string) string {
	return strings.Join(ks, string(separator))
}

// Client represents a Redis store client.
type Client struct {
	*redis.Client
	namespace string
}

// Config represents Redis configuration.
type Config struct {
	config.Redis
	Namespace []string
}

// New returns a new initialized Redis store.
func New(conf *Config) *Client {
	return &Client{
		namespace: Key(append(conf.Redis.Namespace, conf.Namespace...)...),
		Client: redis.NewClient(&redis.Options{
			Addr:     conf.Address,
			Password: conf.Password,
			DB:       conf.Database,
		}),
	}
}

// Key constructs the full key for entity identified by ks by prepending the configured namespace and joining ks using the default separator.
func (s *Client) Key(ks ...string) string {
	return Key(append([]string{s.namespace}, ks...)...)
}

type ProtoCmd struct {
	result func() (string, error)
}

func (cmd ProtoCmd) ScanProto(pb proto.Message) error {
	s, err := cmd.result()
	if err != nil {
		return ConvertError(err)
	}
	return UnmarshalProto(s, pb)
}

// GetProto unmarshals protocol buffer message stored under key k in r into pb.
// Note, that GetProto passes k verbatim to the underlying store and hence, k must represent the full key(including namespace etc.).
func GetProto(r redis.Cmdable, k string) *ProtoCmd {
	return &ProtoCmd{r.Get(k).Result}
}

// SetProto marshals protocol buffer message represented by pb and stores it under key k in r.
// Note, that SetProto passes k verbatim to the underlying store and hence, k must represent the full key(including namespace etc.).
func SetProto(r redis.Cmdable, k string, pb proto.Message, expiration time.Duration) (*redis.StatusCmd, error) {
	s, err := MarshalProto(pb)
	if err != nil {
		return nil, err
	}
	return r.Set(k, s, expiration), nil
}

// WatchCmdable is transactional redis.Cmdable.
type WatchCmdable interface {
	redis.Cmdable
	Watch(fn func(*redis.Tx) error, keys ...string) error
}

func FindProto(r WatchCmdable, k string, keyCmd func(...string) string) *ProtoCmd {
	var result func() (string, error)
	if err := r.Watch(func(tx *redis.Tx) error {
		id, err := tx.Get(k).Result()
		if err != nil {
			return err
		}
		result = tx.Get(keyCmd(id)).Result
		return nil
	}, k); err != nil {
		return &ProtoCmd{result: func() (string, error) { return "", err }}
	}
	return &ProtoCmd{result: result}
}

type ProtosCmd struct {
	result func() ([]string, error)
}

func (cmd ProtosCmd) Range(f func() (proto.Message, func() (bool, error))) error {
	ss, err := cmd.result()
	if err != nil {
		return err
	}
	for _, s := range ss {
		pb, cb := f()
		if err := UnmarshalProto(s, pb); err != nil {
			return err
		}
		if ok, err := cb(); err != nil {
			return err
		} else if !ok {
			return nil
		}
	}
	return nil
}

// FindProtos gets protos stored under keys in k.
func FindProtos(r redis.Cmdable, k string, keyCmd func(...string) string) *ProtosCmd {
	return &ProtosCmd{
		result: r.Sort(k, &redis.Sort{
			Alpha: true,
			Get:   []string{keyCmd("*")},
		}).Result,
	}
}
