// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

// Package redis provides a general Redis client and utilities.
package redis

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"runtime/trace"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/redis/go-redis/v9"
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"google.golang.org/protobuf/proto"
)

const (
	// separator is character used to separate the keys.
	separator = ':'

	// DefaultStreamBlockLimit is the duration for which stream blocking operations
	// such as XRead and XReadGroup should block. Note that Redis operations cannot be
	// asynchronously cancelled using context.WithCancel, so long-polling is required.
	DefaultStreamBlockLimit time.Duration = 0
)

var encoding = base64.RawStdEncoding

// WatchCmdable is transactional redis.Cmdable.
type WatchCmdable interface {
	redis.Cmdable
	Watch(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error
}

// MarshalProto marshals pb into printable string.
func MarshalProto(pb proto.Message) (string, error) {
	b, err := proto.Marshal(pb)
	if err != nil {
		return "", errEncode.WithCause(err)
	}
	protosMarshaled.Inc()
	return encoding.EncodeToString(b), nil
}

// UnmarshalProto unmarshals string returned from MarshalProto into pb.
func UnmarshalProto(s string, pb proto.Message) error {
	b, err := encoding.DecodeString(s)
	if err != nil {
		return errDecode.WithCause(err)
	}
	if err = proto.Unmarshal(b, pb); err != nil {
		return errDecode.WithCause(err)
	}
	protosUnmarshaled.Inc()
	return nil
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
	Address         string         `name:"address" description:"Address of the Redis server"`
	Password        string         `name:"password" description:"Password of the Redis server"`
	Database        int            `name:"database" description:"Redis database to use"`
	RootNamespace   []string       `name:"namespace" description:"Namespace for Redis keys"`
	PoolSize        int            `name:"pool-size" description:"The maximum number of database connections"`
	IdleTimeout     time.Duration  `name:"idle-timeout" description:"Idle connection timeout"`
	ConnMaxLifetime time.Duration  `name:"conn-max-lifetime" description:"Maximum lifetime of a connection"`
	Failover        FailoverConfig `name:"failover" description:"Redis failover configuration"`
	TLS             struct {
		Require          bool `name:"require" description:"Require TLS"`
		tlsconfig.Client `name:",squash"`
	} `name:"tls"`
	namespace []string
}

func equalsStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Equals checks if the other configuration is equivalent to this.
func (c Config) Equals(other Config) bool {
	return c.Address == other.Address &&
		c.Password == other.Password &&
		c.Database == other.Database &&
		equalsStringSlice(c.RootNamespace, other.RootNamespace) &&
		c.PoolSize == other.PoolSize &&
		c.IdleTimeout == other.IdleTimeout &&
		c.ConnMaxLifetime == other.ConnMaxLifetime &&
		c.Failover.Equals(other.Failover) &&
		c.TLS.Require == other.TLS.Require &&
		c.TLS.Client.Equals(other.TLS.Client)
}

func (c Config) WithNamespace(namespace ...string) *Config {
	deriv := c
	deriv.namespace = namespace
	return &deriv
}

// IsZero returns whether the Redis configuration is empty.
func (c Config) IsZero() bool {
	if c.Failover.Enable {
		return c.Failover.MasterName == "" && len(c.Failover.Addresses) == 0
	}
	return c.Address == ""
}

// FailoverConfig represents Redis failover configuration.
type FailoverConfig struct {
	Enable     bool     `name:"enable" description:"Enable failover using Redis Sentinel"`
	Addresses  []string `name:"addresses" description:"Redis Sentinel server addresses"`
	MasterName string   `name:"master-name" description:"Redis Sentinel master name"`
}

// Equals checks if the other configuration is equivalent to this.
func (c FailoverConfig) Equals(other FailoverConfig) bool {
	return c.Enable == other.Enable &&
		equalsStringSlice(c.Addresses, other.Addresses) &&
		c.MasterName == other.MasterName
}

func (c Config) makeDialer() func(ctx context.Context, network, addr string) (net.Conn, error) {
	var (
		dialer interface {
			DialContext(ctx context.Context, network, addr string) (net.Conn, error)
		} = &net.Dialer{}
		tlsConfigErr error
	)
	if c.TLS.Require {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		dialer = &tls.Dialer{NetDialer: dialer.(*net.Dialer), Config: tlsConfig}
		tlsConfigErr = c.TLS.Client.ApplyTo(tlsConfig)
	}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		defer trace.StartRegion(ctx, "dial redis").End()
		if tlsConfigErr != nil {
			return nil, tlsConfigErr
		}
		conn, err := dialer.DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		return &observableConn{addr: addr, Conn: conn}, nil
	}
}

type logFunc func(context.Context, string, ...any)

func (f logFunc) Printf(ctx context.Context, format string, v ...any) {
	f(ctx, format, v...)
}

func debugLogFunc(ctx context.Context, format string, v ...any) {
	log.FromContext(ctx).WithField("origin", "go-redis").Debugf(format, v...)
}

// newRedisClient returns a Redis client, which connects using correct client type.
func newRedisClient(conf *Config) *redis.Client {
	if conf.Failover.Enable {
		return redis.NewFailoverClient(&redis.FailoverOptions{
			Dialer:          conf.makeDialer(),
			MasterName:      conf.Failover.MasterName,
			SentinelAddrs:   conf.Failover.Addresses,
			Password:        conf.Password,
			DB:              conf.Database,
			PoolSize:        conf.PoolSize,
			ConnMaxIdleTime: conf.IdleTimeout,
			ConnMaxLifetime: conf.ConnMaxLifetime,
		})
	}
	return redis.NewClient(&redis.Options{
		Dialer:          conf.makeDialer(),
		Addr:            conf.Address,
		Password:        conf.Password,
		DB:              conf.Database,
		PoolSize:        conf.PoolSize,
		ConnMaxIdleTime: conf.IdleTimeout,
		ConnMaxLifetime: conf.ConnMaxLifetime,
	})
}

// New returns a new initialized Redis store.
func New(conf *Config) *Client {
	return &Client{
		namespace: Key(append(conf.RootNamespace, conf.namespace...)...),
		Client:    newRedisClient(conf),
	}
}

// Key constructs the full key for entity identified by ks by prepending the configured namespace and joining ks using the default separator.
func (cl *Client) Key(ks ...string) string {
	return Key(append([]string{cl.namespace}, ks...)...)
}

// ProtoCmd is a command, which can unmarshal its result into a protocol buffer.
type ProtoCmd struct {
	result func() (string, error)
}

// ScanProto scans command result into proto.Message pb.
func (cmd ProtoCmd) ScanProto(pb proto.Message) error {
	s, err := cmd.result()
	if err != nil {
		return ConvertError(err)
	}
	return UnmarshalProto(s, pb)
}

// GetProto unmarshals protocol buffer message stored under key k in r into pb.
// Note, that GetProto passes k verbatim to the underlying store and hence, k must represent the full key(including namespace etc.).
func GetProto(ctx context.Context, r redis.Cmdable, k string) *ProtoCmd {
	trace.Logf(ctx, "redis", "get proto from %s", k)
	return &ProtoCmd{r.Get(ctx, k).Result}
}

// SetProto marshals protocol buffer message represented by pb and stores it under key k in r.
// Note, that SetProto passes k verbatim to the underlying store and hence, k must represent the full key(including namespace etc.).
func SetProto(ctx context.Context, r redis.Cmdable, k string, pb proto.Message, expiration time.Duration) (*redis.StatusCmd, error) {
	s, err := MarshalProto(pb)
	if err != nil {
		return nil, err
	}
	trace.Logf(ctx, "redis", "set proto to %q", k)
	return r.Set(ctx, k, s, expiration), nil
}

// FindProto finds the protocol buffer stored under the key stored under k.
// The external key is constructed using keyCmd.
func FindProto(ctx context.Context, r WatchCmdable, k string, keyCmd func(string) (string, error)) *ProtoCmd {
	defer trace.StartRegion(ctx, "find proto").End()
	var result func() (string, error)
	if err := r.Watch(ctx, func(tx *redis.Tx) error {
		trace.Logf(ctx, "redis", "get key reference from %q", k)
		id, err := tx.Get(ctx, k).Result()
		if err != nil {
			return err
		}
		ik, err := keyCmd(id)
		if err != nil {
			return err
		}
		trace.Logf(ctx, "redis", "get proto from %q", ik)
		result = tx.Get(ctx, ik).Result
		return nil
	}, k); err != nil {
		return &ProtoCmd{result: func() (string, error) { return "", err }}
	}
	return &ProtoCmd{result: result}
}

type stringSliceCmd struct {
	result func() ([]string, error)
}

// ProtosCmd is a command, which can unmarshal its result into multiple protocol buffers.
type ProtosCmd stringSliceCmd

// Range ranges over command result and unmarshals it into a protocol buffer.
// f must return a new empty proto.Message of the type expected to be present in the command.
// The function returned by f will be called after the commands result is unmarshaled into the message returned by f.
// If both the function returned by f and the message are nil, the entry is skipped.
func (cmd ProtosCmd) Range(f func() (proto.Message, func() (bool, error))) error {
	ss, err := cmd.result()
	if err != nil {
		return err
	}
	for _, s := range ss {
		if s == "" {
			continue
		}

		pb, cb := f()
		if pb == nil && cb == nil {
			continue
		}
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

// ProtosWithKeysCmd is a command, which can unmarshal its result into multiple protocol buffers given a key.
type ProtosWithKeysCmd stringSliceCmd

// Range ranges over command result and unmarshals it into a protocol buffer.
// f must return a new empty proto.Message of the type expected to be present in the command given the key.
// The function returned by f will be called after the commands result is unmarshaled into the message returned by f.
// If both the function returned by f and the message are nil, the entry is skipped.
func (cmd ProtosWithKeysCmd) Range(f func(string) (proto.Message, func() (bool, error))) error {
	ss, err := cmd.result()
	if err != nil {
		return err
	}
	if len(ss)%2 != 0 {
		panic(fmt.Sprintf("odd slice length: %d", len(ss)))
	}
	for i := 0; i < len(ss); i += 2 {
		if ss[i+1] == "" {
			continue
		}

		pb, cb := f(ss[i])
		if pb == nil && cb == nil {
			continue
		}
		if err := UnmarshalProto(ss[i+1], pb); err != nil {
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

type redisSort struct {
	*redis.Sort
}

// FindProtosOption is an option for the FindProtos query.
type FindProtosOption func(redisSort)

// FindProtosSorted ensures that entries are sorted. If alpha is true, lexicographical sorting is used, otherwise - numerical.
func FindProtosSorted(alpha bool) FindProtosOption {
	return func(s redisSort) {
		s.Alpha = alpha
		s.By = ""
	}
}

// FindProtosWithOffsetAndCount changes the offset and the limit of the query.
func FindProtosWithOffsetAndCount(offset, count int64) FindProtosOption {
	return func(s redisSort) {
		s.Offset, s.Count = offset, count
	}
}

func findProtos(ctx context.Context, r redis.Cmdable, k string, keyCmd func(string) string, opts ...FindProtosOption) stringSliceCmd {
	getPattern := keyCmd("*")
	s := &redis.Sort{
		Get: []string{getPattern},
		By:  "nosort", // see https://redis.io/commands/sort#skip-sorting-the-elements
	}
	for _, opt := range opts {
		opt(redisSort{s})
	}
	trace.Logf(ctx, "redis", "find %q protos from %q", getPattern, k)
	return stringSliceCmd{
		result: r.Sort(ctx, k, s).Result,
	}
}

// FindProtos gets protos stored under keys in k.
func FindProtos(ctx context.Context, r redis.Cmdable, k string, keyCmd func(string) string, opts ...FindProtosOption) ProtosCmd {
	return ProtosCmd(findProtos(ctx, r, k, keyCmd, opts...))
}

// FindProtosWithKeys gets protos stored under keys in k including the keys.
func FindProtosWithKeys(ctx context.Context, r redis.Cmdable, k string, keyCmd func(string) string, opts ...FindProtosOption) ProtosWithKeysCmd {
	return ProtosWithKeysCmd(findProtos(ctx, r, k, keyCmd, append([]FindProtosOption{func(s redisSort) { s.Get = append([]string{"#"}, s.Get...) }}, opts...)...))
}

// ListProtos gets list of protos stored under key k.
func ListProtos(ctx context.Context, r redis.Cmdable, k string) ProtosCmd {
	trace.Logf(ctx, "redis", "list protos from %q", k)
	return ProtosCmd{
		result: r.LRange(ctx, k, 0, -1).Result,
	}
}

const (
	payloadKey = "payload"
	replaceKey = "replace"
	startAtKey = "start_at"
	nextAtKey  = "next_at"
)

// InputTaskKey returns the subkey of k, where input tasks are stored.
func InputTaskKey(k string) string {
	return Key(k, "input")
}

// ReadyTaskKey returns the subkey of k, where ready tasks are stored.
func ReadyTaskKey(k string) string {
	return Key(k, "ready")
}

// WaitingTaskKey returns the subkey of k, where waiting tasks are stored.
func WaitingTaskKey(k string) string {
	return Key(k, "waiting")
}

// IsConsumerGroupExistsErr returns true if error represents the redis BUSYGROUP error.
func IsConsumerGroupExistsErr(err error) bool {
	return err != nil && err.Error() == "BUSYGROUP Consumer Group name already exists"
}

// initTaskGroup initializes the task group for streams at InputTaskKey(k) and ReadyTaskKey(k).
// It must be called before all other task-related functions at subkeys of k.
func initTaskGroup(ctx context.Context, r redis.Cmdable, group, k string) error {
	_, err := r.Pipelined(ctx, func(p redis.Pipeliner) error {
		p.XGroupCreateMkStream(ctx, InputTaskKey(k), group, "$")
		p.XGroupCreateMkStream(ctx, ReadyTaskKey(k), group, "$")
		return nil
	})
	if IsConsumerGroupExistsErr(err) {
		return nil
	}
	return ConvertError(err)
}

// addTask adds a task identified by payload with timestamp startAt to the stream at InputTaskKey(k).
// maxLen is the approximate length of the stream, to which it may be trimmed.
func addTask(ctx context.Context, r redis.Cmdable, k string, maxLen int64, payload string, startAt time.Time, replace bool) error {
	m := make(map[string]any, 2)
	m[payloadKey] = payload
	if replace {
		m[replaceKey] = replace
	}
	if !startAt.IsZero() {
		m[startAtKey] = startAt.UnixNano()
	}
	return ConvertError(r.XAdd(ctx, &redis.XAddArgs{
		Stream: InputTaskKey(k),
		MaxLen: maxLen,
		Approx: true,
		Values: m,
	}).Err())
}

func parseTime(s string) (time.Time, error) {
	nsec, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, int64(nsec)).UTC(), nil
}

// dispatchTask dispatches tasks for the callers of popTask. At least one dispatcher is required in order to use popTask.
// The tasks are moved from InputTaskKey(k) to WaitingTaskKey(k). Once the task should be dispatched, it is moved form WaitingTaskKey(k) to ReadyTaskKey(k).
// group is the consumer group name.
// consumer is the consumer ID.
// maxLen represents the maximum size of the streams used for dispatching.
// minIdleTime is used for automatic task reclaiming. Only tasks older than minIdleTime will be redispatched from the input stream to the waiting stream.
func dispatchTask(
	ctx context.Context,
	r redis.Cmdable,
	group, consumer string,
	maxLen int64,
	k string,
	blockLimit time.Duration,
) error {
	var (
		readyStream   = ReadyTaskKey(k)
		inputStream   = InputTaskKey(k)
		waitingStream = WaitingTaskKey(k)
	)
	for {
		ret, err := dispatchTaskScript.Run(
			ctx,
			r,
			[]string{readyStream, inputStream, waitingStream},
			group,
			consumer,
			time.Now().UnixNano(),
			maxLen,
		).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return ConvertError(err)
		}

		block := blockLimit
		if ret != nil {
			s, ok := ret.(string)
			if !ok {
				return errInvalidKeyValueType.WithAttributes("key", nextAtKey).WithCause(err)
			}
			nextAt, err := parseTime(s)
			if err != nil {
				return errInvalidKeyValueType.WithAttributes("key", nextAtKey).WithCause(err)
			}
			if nextAt.IsZero() {
				block = -1
			} else {
				now := time.Now()
				if nextAt.Before(now) {
					continue
				}
				// If we have a task that we may dispatch into the future, we will block the
				// input stream only for the duration between the current time and that future
				// time.
				if d := nextAt.Sub(now); block == 0 || d < block {
					block = d
				}
			}
		}

		_, err = r.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{inputStream, ">"},
			Count:    1,
			Block:    block,
		}).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return ConvertError(err)
		}
	}
}

// popTask calls f on the most recent task in the queue, for which timestamp is in range [0, time.Now()] or blocks until such is available or context is done.
// group is the consumer group name.
// consumer is the consumer group ID.
// ReadyTaskKey(k) is the keys to pop from.
// Pipeline is executed even if f returns an error.
// Tasks are acked only if f returns without error.
func popTask(
	ctx context.Context,
	r redis.Cmdable,
	group, consumer string,
	f func(p redis.Pipeliner, payload string, startAt time.Time) error,
	k string,
	blockLimit time.Duration,
) (err error) {
	readyStream := ReadyTaskKey(k)

	processMessage := func(message redis.XMessage) error {
		fields := make(map[string]string, len(message.Values))
		for k, v := range message.Values {
			val, ok := v.(string)
			if !ok {
				panic(fmt.Sprintf("invalid field type %T", v))
			}
			fields[k] = val
		}

		var startAt time.Time
		if s, ok := fields[startAtKey]; ok {
			startAt, err = parseTime(s)
			if err != nil {
				return errInvalidKeyValueType.WithAttributes("key", startAtKey).WithCause(err)
			}
		}

		p := r.Pipeline()
		defer func() {
			// Ensure pipeline is executed even if f fails.
			_, pErr := p.Exec(ctx)
			if err == nil && pErr != nil {
				err = ConvertError(pErr)
			}
		}()

		if err = f(p, fields[payloadKey], startAt); err != nil {
			return err
		}

		p.XAck(ctx, readyStream, group, message.ID)
		p.XDel(ctx, readyStream, message.ID)

		return nil
	}

	var xs []redis.XStream
	for len(xs) == 0 {
		xs, err = r.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{readyStream, ">"},
			Count:    1,
			Block:    blockLimit,
		}).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return ConvertError(err)
		}
	}

	for _, x := range xs {
		for _, message := range x.Messages {
			if err := processMessage(message); err != nil {
				return err
			}
		}
	}

	return nil
}

// TaskQueue is a task queue.
type TaskQueue struct {
	Redis            WatchCmdable
	MaxLen           int64
	Group            string
	Key              string
	StreamBlockLimit time.Duration

	consumerIDs sync.Map // map[string]struct{} of all used consumer ids
}

// Init initializes the task queue.
// It must be called at least once before using the queue.
func (q *TaskQueue) Init(ctx context.Context) error {
	return initTaskGroup(ctx, q.Redis, q.Group, q.Key)
}

// Close closes the TaskQueue.
func (q *TaskQueue) Close(ctx context.Context) error {
	_, err := q.Redis.Pipelined(ctx, func(p redis.Pipeliner) error {
		q.consumerIDs.Range(func(k, v any) bool {
			p.XGroupDelConsumer(ctx, InputTaskKey(q.Key), q.Group, k.(string))
			p.XGroupDelConsumer(ctx, ReadyTaskKey(q.Key), q.Group, k.(string))
			return true
		})
		return nil
	})
	return ConvertError(err)
}

// Add adds a task s to the queue with a timestamp startAt.
func (q *TaskQueue) Add(ctx context.Context, r redis.Cmdable, s string, startAt time.Time, replace bool) error {
	if r == nil {
		r = q.Redis
	}
	return addTask(ctx, r, q.Key, q.MaxLen, s, startAt, replace)
}

// Dispatch dispatches the tasks of the queue. It will continue to run until the context is done.
// consumerID is used to identify the consumer and should be unique for all concurrent calls to Dispatch.
func (q *TaskQueue) Dispatch(ctx context.Context, consumerID string, r redis.Cmdable) error {
	q.consumerIDs.LoadOrStore(consumerID, struct{}{})
	if r == nil {
		r = q.Redis
	}
	return dispatchTask(ctx, r, q.Group, consumerID, q.MaxLen, q.Key, q.StreamBlockLimit)
}

// Pop calls f on the most recent task in the queue, for which timestamp is in range [0, time.Now()],
// if such is available, otherwise it blocks until it is or context is done.
// Pipeline is executed even if f returns an error.
// consumerID is used to identify the consumer and should be unique for all concurrent calls to Pop.
func (q *TaskQueue) Pop(ctx context.Context, consumerID string, r redis.Cmdable, f func(redis.Pipeliner, string, time.Time) error) error {
	q.consumerIDs.LoadOrStore(consumerID, struct{}{})
	if r == nil {
		r = q.Redis
	}
	return popTask(ctx, r, q.Group, consumerID, f, q.Key, q.StreamBlockLimit)
}

var deduplicateProtosScript = redis.NewScript(`local exp = table.remove(ARGV, 1)
local limit = tonumber(table.remove(ARGV, 1))
local ok = redis.call('set', KEYS[1], '', 'px', exp, 'nx')

if #ARGV > 0 then
	redis.call('rpush', KEYS[2], unpack(ARGV))
	local ttl = redis.call('pttl', KEYS[1])
	redis.call('pexpire', KEYS[2], ttl)
	if limit > 0 then
		redis.call('ltrim', KEYS[2], -limit, -1)
	end
end
if ok then
	return 1
else
	return 0
end`)

// LockKey returns the key lock for k is stored under.
func LockKey(k string) string {
	return Key(k, "lock")
}

// ListKey returns the key list for k is stored under.
func ListKey(k string) string {
	return Key(k, "list")
}

func milliseconds(d time.Duration) int64 {
	ms := d.Milliseconds()
	if ms == 0 && d > 0 {
		return 1
	}
	return ms
}

// DeduplicateProtos deduplicates protos using key k. It stores a lock at LockKey(k)
// and the list of collected protos at ListKey(k).
// If the number of protos exceeds limit, the messages are trimmed from the start of the list.
func DeduplicateProtos(
	ctx context.Context, r redis.Scripter, k string, window time.Duration, limit int, msgs ...proto.Message,
) (bool, error) {
	args := make([]any, 0, 2+len(msgs))
	args = append(args, milliseconds(window))
	args = append(args, limit)
	if n := len(msgs) - limit; n > 0 {
		msgs = msgs[n:]
	}

	for _, msg := range msgs {
		s, err := MarshalProto(msg)
		if err != nil {
			return false, err
		}
		args = append(args, s)
	}
	res, err := deduplicateProtosScript.Run(ctx, r, []string{LockKey(k), ListKey(k)}, args...).Int64()
	if err != nil {
		return false, ConvertError(err)
	}
	return res == 1, nil
}

// NOTE: Time stops in lua scripts and expired keys stay available.

// lockMutexScript attempts to acquire mutex lock.
// It returns 0 if lock is acquired or active locks TTL otherwise.
var lockMutexScript = redis.NewScript(`local pttl = redis.call('pttl', KEYS[1])
if pttl > 0 then
	return pttl
else
	redis.call('del', KEYS[2])
	redis.call('set', KEYS[1], ARGV[1], 'px', ARGV[2])
	return 0
end`)

// takeMutexLockScript attempts to take over the lock from previous caller.
// It returns 1 if lock is acquired and 0 otherwise
var takeMutexLockScript = redis.NewScript(`if redis.call('get', KEYS[1]) == ARGV[1] then
	redis.call('del', KEYS[2])
	redis.call('set', KEYS[1], ARGV[3], 'px', ARGV[2])
	return 1
else
	return 0
end`)

// unlockMutexLockScript unlocks the mutex lock.
var unlockMutexScript = redis.NewScript(`if redis.call('get', KEYS[1]) == ARGV[1] then
	redis.call('lpush', KEYS[2], ARGV[1])
	redis.call('pexpire', KEYS[1], ARGV[2])
	redis.call('pexpire', KEYS[2], ARGV[2])
end
return redis.status_reply('OK')`)

// LockMutex locks the value stored at k with a mutex with identifier id.
// It stores the lock at LockKey(k) and list at ListKey(k).
func LockMutex(ctx context.Context, r redis.Cmdable, k, id string, expiration time.Duration) error {
	defer trace.StartRegion(ctx, "lock mutex").End()

	var hasDeadline bool
	dl, ok := ctx.Deadline()
	if ok {
		hasDeadline = !dl.IsZero()
	}

	lockKey := LockKey(k)
	listKey := ListKey(k)
	expMS := milliseconds(expiration)
	for {
		ttlMS, err := lockMutexScript.Run(ctx, r, []string{lockKey, listKey}, id, expMS).Int64()
		if err != nil {
			return ConvertError(err)
		}
		if ttlMS < 0 {
			panic(fmt.Errorf("negative TTL returned: %d ms", ttlMS))
		}
		if ttlMS == 0 {
			return nil
		}

		timeout := time.Duration(ttlMS) * time.Millisecond
		if hasDeadline {
			until := time.Until(dl)
			if until < timeout {
				timeout = until
			}
		}
		popRes, err := r.BLPop(ctx, timeout, listKey).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return ConvertError(err)
		}
		select {
		case <-ctx.Done():
			if errors.Is(err, redis.Nil) {
				return ctx.Err()
			}
			// Pass the lock to next caller.
			if err := unlockMutexScript.Run(ctx, r, []string{lockKey, listKey}, popRes[1], expMS).Err(); err != nil {
				log.FromContext(ctx).WithError(ConvertError(err)).Error("Failed to pass mutex to next caller")
			}
			return ctx.Err()
		default:
		}
		if errors.Is(err, redis.Nil) {
			continue
		}

		// Attempt to take over the lock from previous caller.
		v, err := takeMutexLockScript.Run(ctx, r, []string{lockKey, listKey}, popRes[1], expMS, id).Int64()
		if err != nil {
			return ConvertError(err)
		}
		if v == 1 {
			return nil
		}
	}
}

// UnlockMutex unlocks the key k with identifier id.
func UnlockMutex(ctx context.Context, r redis.Scripter, k, id string, expiration time.Duration) error {
	defer trace.StartRegion(ctx, "unlock mutex").End()
	if err := unlockMutexScript.Run(ctx, r, []string{LockKey(k), ListKey(k)}, id, milliseconds(expiration)).Err(); err != nil {
		return ConvertError(err)
	}
	return nil
}

// InitMutex initializes the mutex scripts at r.
// InitMutex must be called before mutex functionality is used in a transaction or pipeline.
func InitMutex(ctx context.Context, r redis.Scripter) error {
	if err := lockMutexScript.Load(ctx, r).Err(); err != nil {
		return ConvertError(err)
	}
	if err := takeMutexLockScript.Load(ctx, r).Err(); err != nil {
		return ConvertError(err)
	}
	if err := unlockMutexScript.Load(ctx, r).Err(); err != nil {
		return ConvertError(err)
	}
	return nil
}

// LockedWatch locks the key k with a mutex, watches key k and executes f in a transaction.
// k is unlocked after f returns.
func LockedWatch(ctx context.Context, r WatchCmdable, k, id string, expiration time.Duration, f func(*redis.Tx) error) error {
	defer trace.StartRegion(ctx, "locked watch").End()
	if err := LockMutex(ctx, r, k, id, expiration); err != nil {
		return err
	}
	defer func() {
		if err := UnlockMutex(ctx, r, k, id, expiration); err != nil {
			log.FromContext(ctx).WithField("key", k).WithError(err).Error("Failed to unlock mutex")
		}
	}()
	if err := r.Watch(ctx, f, k); err != nil {
		return ConvertError(err)
	}
	return nil
}

// RangeStreams sequentially iterates over all non-acknowledged messages in streams calling f with at most count
// messages. f must acknowledge the messages which have been processed.
// RangeStreams assumes that within its lifetime it is the only consumer within group group using ID id.
// RangeStreams iterates over all pending messages, which have been idle for at least minIdle milliseconds first.
func RangeStreams(
	ctx context.Context,
	r redis.Cmdable,
	group, id string,
	count int64,
	minIdle time.Duration,
	f func(string, func(...string) error, ...redis.XMessage) error,
	streams ...string,
) error {
	makeAck := func(stream string) func(...string) error {
		return func(ids ...string) error {
			if len(ids) == 0 {
				return nil
			}
			_, err := r.Pipelined(ctx, func(p redis.Pipeliner) error {
				// NOTE: Both calls below copy contents of ids internally.
				p.XAck(ctx, stream, group, ids...)
				p.XDel(ctx, stream, ids...)
				return nil
			})
			if err != nil {
				return ConvertError(err)
			}
			return nil
		}
	}

	for _, stream := range streams {
		for start := "-"; start != "0-0"; {
			var err error
			_, start, err = r.XAutoClaimJustID(ctx, &redis.XAutoClaimArgs{
				Stream:   stream,
				Group:    group,
				Consumer: id,
				MinIdle:  minIdle,
				Start:    start,
				Count:    count,
			}).Result()
			if err != nil {
				return ConvertError(err)
			}
		}
	}

	streamCount := len(streams)
	args := make([]string, 2*streamCount)
	streamsArg := args[:streamCount]
	idsArg := args[streamCount:]
	for i := range streams {
		streamsArg[i], idsArg[i] = streams[i], "0"
	}

	finishedOld, block := false, time.Duration(-1)
	for {
		rets, err := r.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: id,
			Streams:  args,
			Count:    count,
			Block:    block,
		}).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return nil
			}
			return ConvertError(err)
		}

		cont := false
		for i, ret := range rets {
			n := int64(len(ret.Messages))
			if n == 0 {
				continue
			}
			cont = cont || n == count
			idsArg[i] = ret.Messages[len(ret.Messages)-1].ID
			if err := f(ret.Stream, makeAck(ret.Stream), ret.Messages...); err != nil {
				return err
			}
		}

		switch {
		case cont:
			// At least one stream has returned `count` messages.
		case finishedOld:
			// All streams have returned less than `count` messages,
			// and we have already processed all of the old and new messages.
			return nil
		default: // !cont && !finishedOld
			// All streams have returned less than `count` messages,
			// and we have processed all of the old messages.
			finishedOld, block = true, minIdle
			for i := range streams {
				idsArg[i] = ">"
			}
		}
	}
}

// GenerateLockerID generates a unique locker ID to be used with a Redis mutex.
func GenerateLockerID() (string, error) {
	lockID, err := ulid.New(ulid.Now(), rand.Reader)
	if err != nil {
		return "", err
	}
	return lockID.String(), nil
}

func init() {
	redis.SetLogger(logFunc(debugLogFunc))
}
