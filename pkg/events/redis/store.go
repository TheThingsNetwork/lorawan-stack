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
	"crypto/sha1"
	"encoding/base32"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// PubSubStore is a PubSub with historical event storage.
type PubSubStore struct {
	*PubSub

	historyTTL                time.Duration
	entityHistoryCount        int
	entityHistoryTTL          time.Duration
	correlationIDHistoryCount int
}

func (ps *PubSubStore) eventDataKey(_ context.Context, uid string) string {
	return ps.client.Key("event_data", uid)
}

// eventIndexKey returns the Redis for indexing events with the given correlationID.
// Since correlation IDs are user input (and can therefore contain characters that
// we may not want in Redis keys) we hash them.
func (ps *PubSubStore) eventIndexKey(_ context.Context, correlationID string) string {
	h := sha1.New()
	h.Write([]byte(correlationID))
	return ps.client.Key("event_index", base32.StdEncoding.EncodeToString(h.Sum(nil)))
}

func (ps *PubSubStore) eventStream(ctx context.Context, ids *ttnpb.EntityIdentifiers) string {
	if ids == nil {
		return ps.client.Key("event_stream")
	}
	return ps.client.Key("event_stream", ids.EntityType(), unique.ID(ctx, ids))
}

func (ps *PubSubStore) storeEvent(ctx context.Context, evt events.Event) error {
	b, err := encodeEventData(evt)
	if err != nil {
		return err
	}
	_, err = ps.client.TxPipelined(ctx, func(tx redis.Pipeliner) error {
		tx.Set(ps.ctx, ps.eventDataKey(evt.Context(), evt.UniqueID()), b, ps.historyTTL)
		for _, cid := range evt.CorrelationIDs() {
			key := ps.eventIndexKey(evt.Context(), cid)
			tx.LPush(ps.ctx, key, evt.UniqueID())
			tx.LTrim(ps.ctx, key, 0, int64(ps.correlationIDHistoryCount))
			tx.Expire(ps.ctx, key, ps.historyTTL)
		}
		return nil
	})
	return ttnredis.ConvertError(err)
}

// loadEventData loads event data for every event in evts (by UniqueID)
// and set additional fields in the events.
func (ps *PubSubStore) loadEventData(ctx context.Context, evts ...*ttnpb.Event) error {
	if len(evts) == 0 {
		return nil
	}
	keys := make([]string, 0, len(evts))
	for _, evt := range evts {
		keys = append(keys, ps.eventDataKey(ctx, evt.UniqueId))
	}
	bs, err := ps.client.MGet(ctx, keys...).Result()
	if err != nil {
		return ttnredis.ConvertError(err)
	}
	for i, b := range bs {
		switch b := b.(type) {
		case nil:
			continue // Event data deleted/expired.
		case string:
			if err = decodeEventData(b, evts[i]); err != nil {
				log.FromContext(ctx).WithError(err).Warnf("Failed to decode event payload")
				continue
			}
		default:
			log.FromContext(ctx).Warnf("Invalid %T element in event payloads", b)
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
	var evtPB ttnpb.Event
	if err = decodeEventData(data, &evtPB); err != nil {
		return nil, err
	}
	return &evtPB, nil
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

func (ps *PubSubStore) tailStream(ctx context.Context, names []string, ids *ttnpb.EntityIdentifiers, start string, tail int) (events []*ttnpb.Event, nextStart string, err error) {
	matchNames := xMessageHasEventName(names...)
	if tail > 0 {
		msgs, err := ps.client.XRevRangeN(ctx, ps.eventStream(ctx, ids), "+", start, int64(tail)).Result()
		if err != nil {
			return nil, "", ttnredis.ConvertError(err)
		}
		if len(msgs) == 0 {
			return nil, start, nil
		}
		events = eventsFromXMessages(msgs, matchNames)
		reverseEvents(events)
		nextStart = msgs[0].ID
	} else {
		msgs, err := ps.client.XRange(ctx, ps.eventStream(ctx, ids), start, "+").Result()
		if err != nil {
			return nil, "", ttnredis.ConvertError(err)
		}
		if len(msgs) == 0 {
			return nil, start, nil
		}
		events = eventsFromXMessages(msgs, matchNames)
		nextStart = msgs[len(msgs)-1].ID
	}
	if err = ps.loadEventData(ctx, events...); err != nil {
		return nil, "", err
	}
	return events, nextStart, nil
}

// FetchHistory fetches the tail (optional) of historical events matching the given names (optional) and identifiers (mandatory) after the given time (optional).
func (ps *PubSubStore) FetchHistory(ctx context.Context, names []string, ids []*ttnpb.EntityIdentifiers, after *time.Time, tail int) ([]events.Event, error) {
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
		start = strconv.FormatInt(after.Add(-1*time.Second).UnixNano()/1_000_000, 10)
	}
	var evts []events.Event
	for _, id := range ids {
		evtPBs, _, err := ps.tailStream(ctx, names, id, start, tail)
		if err != nil {
			return nil, err
		}
		for _, evtPB := range evtPBs {
			if after != nil {
				if evtTime := ttnpb.StdTime(evtPB.GetTime()); evtTime != nil && !evtTime.Truncate(time.Millisecond).After(*after) {
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

// SubscribeWithHistory is like FetchHistory, but after fetching historical events, this continues sending live events until the context is done.
func (ps *PubSubStore) SubscribeWithHistory(ctx context.Context, names []string, ids []*ttnpb.EntityIdentifiers, after *time.Time, tail int, hdl events.Handler) error {
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
		start = strconv.FormatInt(after.Add(-1*time.Second).UnixNano()/1_000_000, 10)
	}

	type streamState struct {
		id     *ttnpb.EntityIdentifiers
		stream string
		start  string
	}
	state := make([]*streamState, len(ids))
	for i, id := range ids {
		state[i] = &streamState{
			id:     id,
			stream: ps.eventStream(ctx, id),
			start:  start,
		}
	}
	getState := func(stream string) *streamState {
		for _, state := range state {
			if state.stream == stream {
				return state
			}
		}
		return nil
	}

	for _, s := range state {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		evtPBs, nextStart, err := ps.tailStream(ctx, names, s.id, s.start, tail)
		if err != nil {
			return err
		}
		for _, evtPB := range evtPBs {
			if after != nil {
				if evtTime := ttnpb.StdTime(evtPB.GetTime()); evtTime != nil && !evtTime.Truncate(time.Millisecond).After(*after) {
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

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			args := &redis.XReadArgs{
				Count: 10,
				Block: time.Second,
			}
			for _, s := range state {
				args.Streams = append(args.Streams, s.stream)
			}
			for _, s := range state {
				args.Streams = append(args.Streams, s.start)
			}

			streams, err := ps.client.XRead(ctx, args).Result()
			if err != nil && err != redis.Nil {
				return ttnredis.ConvertError(err)
			}
			for _, stream := range streams {
				if len(stream.Messages) == 0 {
					continue
				}
				state := getState(stream.Stream)
				if state == nil {
					continue
				}
				state.start = stream.Messages[len(stream.Messages)-1].ID
				evtPBs := eventsFromXMessages(stream.Messages, matchNames)
				if len(evtPBs) == 0 {
					continue
				}
				if err = ps.loadEventData(ctx, evtPBs...); err != nil {
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
	if err = ps.loadEventData(ctx, evtPBs...); err != nil {
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

// Publish an event to Redis.
func (ps *PubSubStore) Publish(evt events.Event) {
	logger := log.FromContext(ps.ctx)

	if err := ps.storeEvent(ps.ctx, evt); err != nil {
		logger.WithError(err).Warn("Failed to store event")
		return
	}

	m := metaEncodingPrefix + strings.Join([]string{
		evt.UniqueID(),
	}, " ")

	_, err := ps.client.Pipelined(ps.ctx, func(tx redis.Pipeliner) error {
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
				Stream:       eventStream,
				MaxLenApprox: int64(ps.entityHistoryCount),
				Values:       streamValues,
			})
			tx.Expire(ps.ctx, eventStream, ps.entityHistoryTTL)
			if devID := id.GetDeviceIds(); devID != nil && definition != nil && definition.PropagateToParent() {
				eventStream := ps.eventStream(evt.Context(), devID.ApplicationIds.GetEntityIdentifiers())
				tx.XAdd(ps.ctx, &redis.XAddArgs{
					Stream:       eventStream,
					MaxLenApprox: int64(ps.entityHistoryCount),
					Values:       streamValues,
				})
				tx.Expire(ps.ctx, eventStream, ps.entityHistoryTTL)
			}
		}
		return nil
	})
	if err != nil {
		logger.WithError(err).Warn("Failed to publish event")
	}
}
