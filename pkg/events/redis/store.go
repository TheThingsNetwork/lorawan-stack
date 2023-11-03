// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"encoding/base64"
	"errors"
	"hash/fnv"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

const ttlJitter = 0.01

// PubSubStore is a PubSub with historical event storage.
type PubSubStore struct {
	*PubSub

	taskStarter task.Starter
	publisher   events.Publisher

	historyTTL                time.Duration
	entityHistoryCount        int
	entityHistoryTTL          time.Duration
	correlationIDHistoryCount int
	streamPartitionSize       int
}

func (ps *PubSubStore) eventDataKey(_ context.Context, uid string) string {
	return ps.client.Key("event_data", uid)
}

// eventIndexKey returns the Redis for indexing events with the given correlationID.
// Since correlation IDs are user input (and can therefore contain characters that
// we may not want in Redis keys) we hash them.
func (ps *PubSubStore) eventIndexKey(_ context.Context, correlationID string) string {
	h := fnv.New128a()
	h.Write([]byte(correlationID))
	return ps.client.Key("event_index", base64.RawStdEncoding.EncodeToString(h.Sum(nil)))
}

func (ps *PubSubStore) eventStream(ctx context.Context, ids *ttnpb.EntityIdentifiers) string {
	if ids == nil {
		return ps.client.Key("event_stream")
	}
	return ps.client.Key("event_stream", ids.EntityType(), unique.ID(ctx, ids))
}

func (ps *PubSubStore) storeEvent(
	ctx context.Context, tx redis.Cmdable, evt events.Event, correlationIDKeys map[string][]any,
) error {
	b, err := encodeEventData(evt)
	if err != nil {
		return err
	}
	ttl := random.Jitter(ps.historyTTL, ttlJitter)
	tx.Set(ctx, ps.eventDataKey(evt.Context(), evt.UniqueID()), b, ttl)
	for _, cid := range evt.CorrelationIds() {
		key := ps.eventIndexKey(evt.Context(), cid)
		correlationIDKeys[key] = append(correlationIDKeys[key], evt.UniqueID())
	}
	return nil
}

// loadEventData loads event data for every event in evts (by UniqueID)
// and set additional fields in the events.
func (ps *PubSubStore) loadEventData(ctx context.Context, cl redis.Cmdable, evts ...*ttnpb.Event) error {
	if len(evts) == 0 {
		return nil
	}
	keys := make([]string, 0, len(evts))
	for _, evt := range evts {
		keys = append(keys, ps.eventDataKey(ctx, evt.UniqueId))
	}
	bs, err := cl.MGet(ctx, keys...).Result()
	if err != nil {
		return ttnredis.ConvertError(err)
	}
	for i, b := range bs {
		switch b := b.(type) {
		case nil:
			continue // Event data deleted/expired.
		case string:
			if err = decodeEventData(b, evts[i]); err != nil {
				log.FromContext(ctx).WithError(err).Warn("Failed to decode event payload")
				continue
			}
		default:
			log.FromContext(ctx).WithField("element", b).Warn("Invalid element in event payloads")
			continue
		}
	}
	return nil
}

// LoadEvent loads an event by its UID.
func (ps *PubSubStore) LoadEvent(ctx context.Context, uid string) (*ttnpb.Event, error) {
	data, err := ps.client.Get(ctx, ps.eventDataKey(ctx, uid)).Result()
	if err != nil {
		return nil, ttnredis.ConvertError(err)
	}
	evtPB := &ttnpb.Event{}
	if err = decodeEventData(data, evtPB); err != nil {
		return nil, err
	}
	return evtPB, nil
}

func xMessageHasEventName(names ...string) func(msg redis.XMessage) bool {
	if len(names) == 0 {
		return func(msg redis.XMessage) bool {
			return true
		}
	}
	nameMap := make(map[string]struct{}, len(names))
	for _, name := range names {
		nameMap[name] = struct{}{}
	}
	return func(msg redis.XMessage) bool {
		name, ok := msg.Values[eventNameKey].(string)
		if !ok {
			return false
		}
		_, match := nameMap[name]
		return match
	}
}

func eventsFromXMessages(msgs []redis.XMessage, match func(msg redis.XMessage) bool) []*ttnpb.Event {
	evts := make([]*ttnpb.Event, 0, len(msgs))
	for _, msg := range msgs {
		if match != nil && !match(msg) {
			continue
		}
		if evt, err := decodeEventMeta(msg.Values); err == nil {
			evts = append(evts, evt)
		}
	}
	return evts
}

func reverseEvents(slice []*ttnpb.Event) {
	for i := 0; i < len(slice)/2; i++ {
		j := len(slice) - i - 1
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func (ps *PubSubStore) tailStream(
	ctx context.Context, names []string, ids *ttnpb.EntityIdentifiers, start string, tail int,
) (eventPBs []*ttnpb.Event, nextStart string, err error) {
	matchNames := xMessageHasEventName(names...)
	if tail > 0 {
		msgs, err := ps.client.XRevRangeN(ctx, ps.eventStream(ctx, ids), "+", start, int64(tail)).Result()
		if err != nil {
			return nil, "", ttnredis.ConvertError(err)
		}
		if len(msgs) == 0 {
			return nil, start, nil
		}
		eventPBs = eventsFromXMessages(msgs, matchNames)
		reverseEvents(eventPBs)
		nextStart = msgs[0].ID
	} else {
		msgs, err := ps.client.XRange(ctx, ps.eventStream(ctx, ids), start, "+").Result()
		if err != nil {
			return nil, "", ttnredis.ConvertError(err)
		}
		if len(msgs) == 0 {
			return nil, start, nil
		}
		eventPBs = eventsFromXMessages(msgs, matchNames)
		nextStart = msgs[len(msgs)-1].ID
	}
	if err = ps.loadEventData(ctx, ps.client, eventPBs...); err != nil {
		return nil, "", err
	}
	return eventPBs, nextStart, nil
}

// FetchHistory fetches the tail (optional) of historical events matching the given
// names (optional) and identifiers (mandatory) after the given time (optional).
func (ps *PubSubStore) FetchHistory(
	ctx context.Context, names []string, ids []*ttnpb.EntityIdentifiers, after *time.Time, tail int,
) ([]events.Event, error) {
	start := "-"
	switch {
	case after == nil:
	case after.IsZero():
		after = nil
	default:
		// Truncate to milliseconds to be consistent with the JSON API.
		afterMS := after.Truncate(time.Millisecond)
		after = &afterMS
		// Account for a clock skew on the Redis server of up to 1 second.
		start = formatStreamTime(after.Add(-1 * time.Second))
	}
	var evts []events.Event
	for _, id := range ids {
		evtPBs, _, err := ps.tailStream(ctx, names, id, start, tail)
		if err != nil {
			return nil, err
		}
		for _, evtPB := range evtPBs {
			if after != nil {
				if evtTime := ttnpb.StdTime(evtPB.GetTime()); evtTime != nil &&
					!evtTime.Truncate(time.Millisecond).After(*after) {
					continue
				}
			}
			evt, err := events.FromProto(evtPB)
			if err != nil {
				return nil, err
			}
			evts = append(evts, evt)
		}
	}
	return evts, nil
}

type streamState struct {
	id     *ttnpb.EntityIdentifiers
	stream string
	start  string
}

func streamPartitionSize(states []*streamState, partitionSize int) int {
	size := partitionSize
	if n := len(states); n < size {
		size = n
	}
	return size
}

func partitionStreamStates(states []*streamState, partitionSize int) [][]*streamState {
	states = slices.Clone(states)
	rand.Shuffle(len(states), func(i, j int) { states[i], states[j] = states[j], states[i] })
	partitionedStates := make([][]*streamState, 0, len(states)/partitionSize+1)
	for len(states) > 0 {
		n := streamPartitionSize(states, partitionSize)
		partitionedStates = append(partitionedStates, states[:n])
		states = states[n:]
	}
	return partitionedStates
}

func createStreamStateMapping(states []*streamState) map[string]*streamState {
	m := make(map[string]*streamState, len(states))
	for _, state := range states {
		m[state.stream] = state
	}
	return m
}

func (ps *PubSubStore) iterateStreamPartition(
	ctx context.Context,
	states []*streamState,
	eventCountLimit int64,
	ch chan<- []redis.XStream,
) error {
	statesByStream := createStreamStateMapping(states)
	for {
		args := &redis.XReadArgs{
			Count: eventCountLimit,
			Block: random.Jitter(8*time.Second, 0.2),
		}
		for _, s := range states {
			args.Streams = append(args.Streams, s.stream)
		}
		for _, s := range states {
			args.Streams = append(args.Streams, s.start)
		}
		streams, err := ps.client.XRead(ctx, args).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return ttnredis.ConvertError(err)
		}
		result := make([]redis.XStream, 0, len(streams))
		for _, stream := range streams {
			if len(stream.Messages) == 0 {
				continue
			}
			state, ok := statesByStream[stream.Stream]
			if !ok {
				continue
			}
			state.start = stream.Messages[len(stream.Messages)-1].ID
			result = append(result, stream)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- result:
		}
	}
}

// SubscribeWithHistory is like FetchHistory, but after fetching historical events,
// this continues sending live events until the context is done.
func (ps *PubSubStore) SubscribeWithHistory(
	ctx context.Context, names []string, ids []*ttnpb.EntityIdentifiers, after *time.Time, tail int, hdl events.Handler,
) (err error) {
	start := "0-0"
	switch {
	case after == nil:
	case after.IsZero():
		after = nil
	default:
		// Truncate to milliseconds to be consistent with the JSON API.
		afterMS := after.Truncate(time.Millisecond)
		after = &afterMS
		// Account for a clock skew on the Redis server of up to 1 second.
		start = formatStreamTime(after.Add(-1 * time.Second))
	}

	states := make([]*streamState, len(ids))
	for i, id := range ids {
		states[i] = &streamState{
			id:     id,
			stream: ps.eventStream(ctx, id),
			start:  start,
		}
	}

	for _, s := range states {
		evtPBs, nextStart, err := ps.tailStream(ctx, names, s.id, s.start, tail)
		if err != nil {
			return err
		}
		for _, evtPB := range evtPBs {
			if after != nil {
				if evtTime := ttnpb.StdTime(evtPB.GetTime()); evtTime != nil &&
					!evtTime.Truncate(time.Millisecond).After(*after) {
					continue
				}
			}
			evt, err := events.FromProto(evtPB)
			if err != nil {
				return err
			}
			hdl.Notify(evt)
		}
		s.start = nextStart
	}

	matchNames := xMessageHasEventName(names...)

	// We expect that the subscriber can ingest `tail` events at a time.
	eventCountLimit := int64(tail)
	switch {
	case eventCountLimit < 8:
		eventCountLimit = 8
	case eventCountLimit > 1024:
		eventCountLimit = 1024
	}

	ch := make(chan []redis.XStream, 1)
	wg := sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := errorcontext.New(ctx)
	for _, states := range partitionStreamStates(states, ps.streamPartitionSize) {
		states := states
		f := func(ctx context.Context) (err error) {
			defer func() { cancel(err) }()
			return ps.iterateStreamPartition(ctx, states, eventCountLimit, ch)
		}
		wg.Add(1)
		ps.taskStarter.StartTask(&task.Config{
			Context: ctx,
			ID:      "events_iterate_partition",
			Func:    f,
			Done:    wg.Done,
			Restart: task.RestartNever,
			Backoff: task.DefaultBackoffConfig,
		})
	}

	defer func() { cancel(err) }()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case streams := <-ch:
			for _, stream := range streams {
				evtPBs := eventsFromXMessages(stream.Messages, matchNames)
				if len(evtPBs) == 0 {
					continue
				}
				if err := ps.loadEventData(ctx, ps.client, evtPBs...); err != nil {
					return err
				}
				for _, evtPB := range evtPBs {
					evt, err := events.FromProto(evtPB)
					if err != nil {
						return err
					}
					hdl.Notify(evt)
				}
			}
		}
	}
}

// FindRelated finds events with matching correlation IDs.
func (ps *PubSubStore) FindRelated(ctx context.Context, correlationID string) ([]events.Event, error) {
	uids, err := ps.client.LRange(ctx, ps.eventIndexKey(ctx, correlationID), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	evtPBs := make([]*ttnpb.Event, len(uids))
	for i, uid := range uids {
		evtPBs[i] = &ttnpb.Event{UniqueId: uid}
	}
	if err = ps.loadEventData(ctx, ps.client, evtPBs...); err != nil {
		return nil, err
	}
	evts := make([]events.Event, 0, len(evtPBs))
	for _, evtPB := range evtPBs {
		if evtPB.Name == "" {
			// loadEventData was supposed to write this, so if it's not present
			// the event was no longer available.
			continue
		}
		evt, err := events.FromProto(evtPB)
		if err != nil {
			return nil, err
		}
		evts = append(evts, evt)
	}
	return evts, nil
}

// Publish implements events.Publisher.
func (ps *PubSubStore) Publish(evs ...events.Event) {
	ps.publisher.Publish(evs...)
}

// publish an event to Redis.
func (ps *PubSubStore) publish(evs ...events.Event) {
	logger := log.FromContext(ps.ctx)

	tx := ps.client.TxPipeline()

	eventStreams := make(map[string]struct{}, len(evs))
	correlationIDKeys := make(map[string][]any, len(evs))
	for _, evt := range evs {
		if err := ps.storeEvent(ps.ctx, tx, evt, correlationIDKeys); err != nil {
			logger.WithError(err).Warn("Failed to store event")
			continue
		}

		m := metaEncodingPrefix + strings.Join([]string{
			evt.UniqueID(),
		}, " ")

		ids := evt.Identifiers()
		if len(ids) == 0 {
			tx.Publish(ps.ctx, ps.eventChannel(evt.Context(), evt.Name(), nil), m)
		}
		definition := events.GetDefinition(evt)
		for _, id := range ids {
			tx.Publish(ps.ctx, ps.eventChannel(evt.Context(), evt.Name(), id), m)
			streamValues, err := encodeEventMeta(evt, id)
			if err != nil {
				logger.WithError(err).Warn("Failed to encode event")
				continue
			}
			eventStream := ps.eventStream(evt.Context(), id)
			tx.XAdd(ps.ctx, &redis.XAddArgs{
				Stream: eventStream,
				MaxLen: int64(ps.entityHistoryCount),
				Approx: true,
				Values: streamValues,
			})
			eventStreams[eventStream] = struct{}{}
			if devID := id.GetDeviceIds(); devID != nil && definition != nil && definition.PropagateToParent() {
				eventStream := ps.eventStream(evt.Context(), devID.ApplicationIds.GetEntityIdentifiers())
				eventStreams[eventStream] = struct{}{}
				tx.XAdd(ps.ctx, &redis.XAddArgs{
					Stream: eventStream,
					MaxLen: int64(ps.entityHistoryCount),
					Approx: true,
					Values: streamValues,
				})
			}
		}
	}

	entityHistoryTTL := random.Jitter(ps.entityHistoryTTL, ttlJitter)
	for eventStream := range eventStreams {
		tx.PExpire(ps.ctx, eventStream, entityHistoryTTL)
	}

	historyTTL := random.Jitter(ps.historyTTL, ttlJitter)
	for correlationIDKey, eventIDs := range correlationIDKeys {
		tx.LPush(ps.ctx, correlationIDKey, eventIDs...)
		tx.LTrim(ps.ctx, correlationIDKey, 0, int64(ps.correlationIDHistoryCount))
		tx.PExpire(ps.ctx, correlationIDKey, historyTTL)
	}

	if err := ps.transactionPool.Publish(ps.ctx, tx); err != nil {
		logger.WithError(err).Warn("Failed to publish transaction")
	}
}

// formatStreamTime constructs the minimal stream ID from the provided timestamp.
// Redis stream identifiers are by default built from the number of milliseconds since
// the UNIX epoch, and a sequence number.
func formatStreamTime(t time.Time) string {
	return strconv.FormatInt(t.UnixNano()/int64(time.Millisecond), 10)
}
