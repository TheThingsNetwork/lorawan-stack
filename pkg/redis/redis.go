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
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gogo/protobuf/proto"
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
)

const (
	// separator is character used to separate the keys.
	separator = ':'
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
	Address       string         `name:"address" description:"Address of the Redis server"`
	Password      string         `name:"password" description:"Password of the Redis server"`
	Database      int            `name:"database" description:"Redis database to use"`
	RootNamespace []string       `name:"namespace" description:"Namespace for Redis keys"`
	PoolSize      int            `name:"pool-size" description:"The maximum number of database connections"`
	Failover      FailoverConfig `name:"failover" description:"Redis failover configuration"`
	TLS           struct {
		Require          bool `name:"require" description:"Require TLS"`
		tlsconfig.Client `name:",squash"`
	} `name:"tls"`
	namespace []string
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

func (c Config) makeDialer() func(ctx context.Context, network, addr string) (net.Conn, error) {
	var (
		tlsConfig    *tls.Config
		tlsConfigErr error
	)
	if c.TLS.Require {
		tlsConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		tlsConfigErr = c.TLS.Client.ApplyTo(tlsConfig)
	}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		var timeout time.Duration
		deadline, ok := ctx.Deadline()
		if ok {
			timeout = time.Until(deadline)
		}
		var (
			conn net.Conn
			err  error
		)
		dialer := &net.Dialer{Timeout: timeout}
		if c.TLS.Require {
			if tlsConfigErr != nil {
				return nil, tlsConfigErr
			}
			conn, err = tls.DialWithDialer(dialer, network, addr, tlsConfig)
			if err != nil {
				return nil, err
			}
		} else {
			conn, err = dialer.Dial(network, addr)
			if err != nil {
				return nil, err
			}
		}
		return &observableConn{addr: addr, Conn: conn}, nil
	}
}

type logFunc func(context.Context, string, ...interface{})

func (f logFunc) Printf(ctx context.Context, format string, v ...interface{}) {
	f(ctx, format, v...)
}

func debugLogFunc(ctx context.Context, format string, v ...interface{}) {
	log.FromContext(ctx).WithField("origin", "go-redis").Debugf(format, v...)
}

// newRedisClient returns a Redis client, which connects using correct client type.
func newRedisClient(conf *Config) *redis.Client {
	if conf.Failover.Enable {
		return redis.NewFailoverClient(&redis.FailoverOptions{
			Dialer:        conf.makeDialer(),
			MasterName:    conf.Failover.MasterName,
			SentinelAddrs: conf.Failover.Addresses,
			Password:      conf.Password,
			DB:            conf.Database,
			PoolSize:      conf.PoolSize,
		})
	}
	return redis.NewClient(&redis.Options{
		Dialer:   conf.makeDialer(),
		Addr:     conf.Address,
		Password: conf.Password,
		DB:       conf.Database,
		PoolSize: conf.PoolSize,
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
	return &ProtoCmd{r.Get(ctx, k).Result}
}

// SetProto marshals protocol buffer message represented by pb and stores it under key k in r.
// Note, that SetProto passes k verbatim to the underlying store and hence, k must represent the full key(including namespace etc.).
func SetProto(ctx context.Context, r redis.Cmdable, k string, pb proto.Message, expiration time.Duration) (*redis.StatusCmd, error) {
	s, err := MarshalProto(pb)
	if err != nil {
		return nil, err
	}
	return r.Set(ctx, k, s, expiration), nil
}

// FindProto finds the protocol buffer stored under the key stored under k.
// The external key is constructed using keyCmd.
func FindProto(ctx context.Context, r WatchCmdable, k string, keyCmd func(string) (string, error)) *ProtoCmd {
	var result func() (string, error)
	if err := r.Watch(ctx, func(tx *redis.Tx) error {
		id, err := tx.Get(ctx, k).Result()
		if err != nil {
			return err
		}
		ik, err := keyCmd(id)
		if err != nil {
			return err
		}
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
	s := &redis.Sort{
		Get: []string{keyCmd("*")},
		By:  "nosort", // see https://redis.io/commands/sort#skip-sorting-the-elements
	}
	for _, opt := range opts {
		opt(redisSort{s})
	}
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
	return ProtosCmd{
		result: r.LRange(ctx, k, 0, -1).Result,
	}
}

type InterfaceSliceCmd struct {
	*redis.Cmd
}

func (cmd InterfaceSliceCmd) Result() ([]interface{}, error) {
	v, err := cmd.Cmd.Result()
	if err != nil {
		return nil, err
	}
	vs, ok := v.([]interface{})
	if !ok {
		return nil, errDecode.New()
	}
	return vs, nil
}

func RunInterfaceSliceScript(ctx context.Context, r Scripter, s *redis.Script, keys []string, args ...interface{}) *InterfaceSliceCmd {
	return &InterfaceSliceCmd{s.Run(ctx, r, keys, args...)}
}

const (
	payloadKey = "payload"
	replaceKey = "replace"
	startAtKey = "start_at"
	nextAtKey  = "next_at"
	lastIDKey  = "last_id"
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
	m := make(map[string]interface{}, 2)
	m[payloadKey] = payload
	if replace {
		m[replaceKey] = replace
	}
	if !startAt.IsZero() {
		m[startAtKey] = startAt.UnixNano()
	}
	return ConvertError(r.XAdd(ctx, &redis.XAddArgs{
		Stream:       InputTaskKey(k),
		MaxLenApprox: maxLen,
		Values:       m,
	}).Err())
}

func parseTime(s string) (time.Time, error) {
	nsec, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, int64(nsec)).UTC(), nil
}

// popTask calls f on the most recent task in the queue, for which timestamp is in range [0, time.Now()] or blocks until such is available or context is done.
// If there are no tasks available for immediate processing, popTask lazily dispatches available tasks for itself and all other callers of popTask.
// group is the consumer group name.
// id is the consumer group ID.
// k is the keys to pop from.
// Pipeline is executed even if f returns an error.
// Tasks are acked only if f returns without error.
func popTask(ctx context.Context, r redis.Cmdable, group, id string, maxLen int64, f func(p redis.Pipeliner, payload string, startAt time.Time) error, k string) (err error) {
	var (
		readyStream   = ReadyTaskKey(k)
		inputStream   = InputTaskKey(k)
		waitingStream = WaitingTaskKey(k)
	)
	for {
		vs, err := RunInterfaceSliceScript(ctx, r, popTaskScript, []string{readyStream, inputStream, waitingStream}, group, id, time.Now().UnixNano(), maxLen).Result()
		if err != nil && err != redis.Nil {
			return ConvertError(err)
		}
		typ, ok := vs[0].(string)
		if !ok {
			panic(fmt.Sprintf("invalid type of entry at index %d of result returned by Redis task pop script: %T", 0, vs[0]))
		}
		var fields map[string]string
		if len(vs) > 1 {
			fields = make(map[string]string, (len(vs)-1)/2)
			for i := 1; i < len(vs); i += 2 {
				k, ok := vs[i].(string)
				if !ok {
					panic(fmt.Sprintf("invalid type of entry at index %d of result returned by Redis task pop script: %T", i, vs[i]))
				}
				v, ok := vs[i+1].(string)
				if !ok {
					panic(fmt.Sprintf("invalid type of entry at index %d of result returned by Redis task pop script: %T", i+1, vs[i+1]))
				}
				fields[k] = v
			}
		}

		switch typ {
		case "ready":
		case "waiting":
			xs, err := r.XPendingExt(ctx, &redis.XPendingExtArgs{
				Stream: readyStream,
				Group:  group,
				Start:  "-",
				End:    "+",
				Count:  1,
			}).Result()
			if err != nil && err != redis.Nil {
				return ConvertError(err)
			}
			if len(xs) > 0 {
				// TODO: XCLAIM and handle (https://github.com/TheThingsNetwork/lorawan-stack/issues/44)
			}

			var block time.Duration
			if s, ok := fields[nextAtKey]; ok {
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
					} else {
						block = nextAt.Sub(now)
					}
				}
			}
			var id string
			if s, ok := fields[lastIDKey]; ok {
				id = s
			} else {
				id = "0-0"
			}
			_, err = r.XRead(ctx, &redis.XReadArgs{
				Streams: []string{inputStream, id},
				Count:   1,
				Block:   block,
			}).Result()
			if err != nil && err != redis.Nil {
				return ConvertError(err)
			}
			continue

		default:
			panic(fmt.Sprintf("unknown result type received `%s`", typ))
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
		if err = f(p, fields["payload"], startAt); err != nil {
			return err
		}
		p.XAck(ctx, readyStream, group, fields["id"])
		p.XDel(ctx, readyStream, fields["id"])
		return nil
	}
}

// TaskQueue is a task queue.
type TaskQueue struct {
	Redis  WatchCmdable
	MaxLen int64
	Group  string
	Key    string

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
		q.consumerIDs.Range(func(k, v interface{}) bool {
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

// Pop calls f on the most recent task in the queue, for which timestamp is in range [0, time.Now()],
// if such is available, otherwise it blocks until it is or context is done.
// Pipeline is executed even if f returns an error.
// consumerID is used to identify the consumer and should be unique for all concurrent calls to Pop.
func (q *TaskQueue) Pop(ctx context.Context, consumerID string, r redis.Cmdable, f func(redis.Pipeliner, string, time.Time) error) error {
	q.consumerIDs.LoadOrStore(consumerID, struct{}{})
	if r == nil {
		r = q.Redis
	}
	return popTask(ctx, r, q.Group, consumerID, q.MaxLen, f, q.Key)
}

// Scripter is redis.scripter.
type Scripter interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd
	ScriptExists(ctx context.Context, hashes ...string) *redis.BoolSliceCmd
	ScriptLoad(ctx context.Context, script string) *redis.StringCmd
}

var deduplicateProtosScript = redis.NewScript(`local exp = ARGV[1]
local ok = redis.call('set', KEYS[1], '', 'px', exp, 'nx')
if #ARGV > 1 then
	table.remove(ARGV, 1)
	redis.call('rpush', KEYS[2], unpack(ARGV))
	redis.call('pexpire', KEYS[2], exp)
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

// DeduplicateProtos deduplicates protos using key k. It stores a lock at LockKey(k) and the list of collected protos at ListKey(k).
func DeduplicateProtos(ctx context.Context, r Scripter, k string, window time.Duration, msgs ...proto.Message) (bool, error) {
	args := make([]interface{}, 0, 1+len(msgs))
	args = append(args, milliseconds(window))
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
		if err != nil && err != redis.Nil {
			return ConvertError(err)
		}
		select {
		case <-ctx.Done():
			if err == redis.Nil {
				return ctx.Err()
			}
			// Pass the lock to next caller.
			if err := unlockMutexScript.Run(ctx, r, []string{lockKey, listKey}, popRes[1], expMS).Err(); err != nil {
				log.FromContext(ctx).WithError(ConvertError(err)).Error("Failed to pass mutex to next caller")
			}
			return ctx.Err()
		default:
		}
		if err == redis.Nil {
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
func UnlockMutex(ctx context.Context, r Scripter, k, id string, expiration time.Duration) error {
	if err := unlockMutexScript.Run(ctx, r, []string{LockKey(k), ListKey(k)}, id, milliseconds(expiration)).Err(); err != nil {
		return ConvertError(err)
	}
	return nil
}

// InitMutex initializes the mutex scripts at r.
// InitMutex must be called before mutex functionality is used in a transaction or pipeline.
func InitMutex(ctx context.Context, r Scripter) error {
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

// XAutoClaim provides a Lua implementation of `XAUTOCLAIM` command introduced in Redis 6.2.0.
func XAutoClaim(ctx context.Context, r Scripter, stream, group, id string, minIdle time.Duration, start string, count int64) ([]redis.XMessage, string, error) {
	var (
		vs  []interface{}
		err error
	)
	vs, err = RunInterfaceSliceScript(ctx, r, xAutoClaimScript, []string{stream}, group, id, minIdle.Milliseconds(), start, count).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, "", nil
		}
		return nil, "", ConvertError(err)
	}

	lastID, ok := vs[0].(string)
	if !ok {
		panic(fmt.Sprintf("invalid type of entry at index %d of result returned by Redis xautoclaim script: %T", 0, vs[0]))
	}

	xis, ok := vs[1].([]interface{})
	if !ok {
		panic(fmt.Sprintf("invalid type of entry at index %d of result returned by Redis xautoclaim script: %T", 1, vs[1]))
	}
	if len(xis) == 0 {
		return nil, "", nil
	}

	xs := make([]redis.XMessage, 0, len(xis))
	for i, xi := range xis {
		xvs, ok := xi.([]interface{})
		if !ok {
			panic(fmt.Sprintf("invalid type of xmessage at index %d of result returned by Redis xautoclaim script: %T", i, xi))
		}

		id, ok := xvs[0].(string)
		if !ok {
			panic(fmt.Sprintf("invalid type of ID field of xmessage at index %d of result returned by Redis xautoclaim script: %T", i, xvs[0]))
		}

		ss, ok := xvs[1].([]interface{})
		if !ok {
			panic(fmt.Sprintf("invalid type of value field of xmessage at index %d of result returned by Redis xautoclaim script: %T", i, xvs[1]))
		}
		if n := len(ss); n%2 != 0 {
			panic(fmt.Sprintf("invalid length of value field of xmessage at index %d of result returned by Redis xautoclaim script: %d", i, n))
		}
		values := make(map[string]interface{}, len(ss))
		for j := 0; j < len(ss); j += 2 {
			k, ok := ss[j].(string)
			if !ok {
				panic(fmt.Sprintf("invalid type of key field value field of xmessage at index %d of result returned by Redis xautoclaim script: %T", i, ss[0]))
			}
			values[k] = ss[j+1]
		}
		xs = append(xs, redis.XMessage{
			ID:     id,
			Values: values,
		})
	}
	return xs, lastID, nil
}

// RangeStreams sequentially iterates over all non-acknowledged messages in streams calling f with at most count messages.
// RangeStreams assumes that within its lifetime it is the only consumer within group group using ID id.
// RangeStreams iterates over all pending messages, which have been idle for at least minIdle milliseconds first.
func RangeStreams(ctx context.Context, r redis.Cmdable, group, id string, count int64, minIdle time.Duration, f func(string, ...redis.XMessage) error, streams ...string) error {
	var ack func(context.Context, string, ...redis.XMessage) error
	{
		ids := make([]string, 0, int(count))
		ack = func(ctx context.Context, stream string, msgs ...redis.XMessage) error {
			ids = ids[:0]
			for _, msg := range msgs {
				ids = append(ids, msg.ID)
			}
			_, err := r.Pipelined(ctx, func(p redis.Pipeliner) error {
				// NOTE: Both calls below copy contents of ids internally.
				p.XAck(ctx, stream, group, ids...)
				p.XDel(ctx, stream, ids...)
				return nil
			})
			return err
		}
	}

	for _, stream := range streams {
		for start := "-"; ; {
			msgs, lastID, err := XAutoClaim(ctx, r, stream, group, id, minIdle, start, count)
			if err != nil {
				return err
			}
			if len(msgs) == 0 {
				break
			}
			if err := f(stream, msgs...); err != nil {
				return err
			}
			if err := ack(ctx, stream, msgs...); err != nil {
				return err
			}
			start = lastID
		}
	}

	streamCount := len(streams)
	streamsArg := make([]string, 2*streamCount)
	idsArg := make([]string, 2*streamCount)
	for i, stream := range streams {
		j := i * 2
		streamsArg[j], streamsArg[j+1], idsArg[j], idsArg[j+1] = stream, stream, "0", ">"
	}

	drainedOld := make(map[string]struct{}, streamCount)
outer:
	for {
		rets, err := r.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: id,
			Streams:  append(streamsArg, idsArg...),
			Count:    count,
			Block:    -1, // do not block
		}).Result()
		if err != nil {
			if err == redis.Nil {
				return nil
			}
			return ConvertError(err)
		}

		for i, ret := range rets {
			if idsArg[i] == "0" && len(ret.Messages) < int(count) {
				drainedOld[ret.Stream] = struct{}{}
			}
			if len(ret.Messages) == 0 {
				if i == len(rets)-1 {
					return nil
				}
				continue
			}

			if err := f(ret.Stream, ret.Messages...); err != nil {
				return err
			}
			if err := ack(ctx, ret.Stream, ret.Messages...); err != nil {
				return err
			}
			if len(ret.Messages) == int(count) {
				continue outer
			}
		}

		streamsArg = streamsArg[:0]
		idsArg = idsArg[:0]
		for _, stream := range streams {
			_, ok := drainedOld[stream]
			if !ok {
				streamsArg = append(streamsArg, stream)
				idsArg = append(idsArg, "0")
			}
			streamsArg = append(streamsArg, stream)
			idsArg = append(idsArg, ">")
		}
	}
}

func init() {
	redis.SetLogger(logFunc(debugLogFunc))
}
