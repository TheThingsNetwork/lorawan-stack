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

// Package events implements event handling through a PubSub interface.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Event interface
type Event interface {
	Context() context.Context
	Name() string
	Time() time.Time
	Identifiers() []*ttnpb.EntityIdentifiers
	Data() interface{}
	CorrelationIDs() []string
	Origin() string
	Caller() string
}

func local(evt Event) *event {
	localEvent, ok := evt.(*event)
	if !ok {
		localEvent = &event{
			ctx: evt.Context(),
			innerEvent: ttnpb.Event{
				Name:           evt.Name(),
				Time:           evt.Time(),
				Identifiers:    evt.Identifiers(),
				CorrelationIDs: evt.CorrelationIDs(),
				Origin:         evt.Origin(),
			},
			data:   evt.Data(),
			caller: evt.Caller(),
		}
	}
	return localEvent
}

type event struct {
	ctx        context.Context
	innerEvent ttnpb.Event
	data       interface{}
	caller     string
}

// IncludeCaller indicates whether the caller of Publish should be included in the event
var IncludeCaller bool

// withCaller returns an event with the Caller field populated, if configured to do so.
// If the original event already had a non-empty Caller, the original event is returned.
func (e *event) withCaller() *event {
	if IncludeCaller && e.caller == "" {
		if _, file, line, ok := runtime.Caller(2); ok {
			split := strings.SplitAfter(file, "lorawan-stack/")
			if len(split) > 1 {
				file = split[1]
			}
			clone := *e
			clone.caller = fmt.Sprintf("%s:%d", file, line)
			return &clone
		}
	}
	return e
}

func (e event) MarshalJSON() ([]byte, error) {
	pb, err := Proto(e)
	if err != nil {
		return nil, err
	}
	return jsonpb.TTN().Marshal(pb)
}

func (e *event) UnmarshalJSON(data []byte) error {
	var pb ttnpb.Event
	err := jsonpb.TTN().Unmarshal(data, &pb)
	if err != nil {
		return err
	}
	fromProto, err := FromProto(&pb)
	if err != nil {
		return err
	}
	evt := fromProto.(*event)
	*e = *evt
	return nil
}

func (e event) Context() context.Context                { return e.ctx }
func (e event) Name() string                            { return e.innerEvent.Name }
func (e event) Time() time.Time                         { return e.innerEvent.Time }
func (e event) Identifiers() []*ttnpb.EntityIdentifiers { return e.innerEvent.Identifiers }
func (e event) Data() interface{}                       { return e.data }
func (e event) CorrelationIDs() []string                { return e.innerEvent.CorrelationIDs }
func (e event) Origin() string                          { return e.innerEvent.Origin }
func (e event) Caller() string                          { return e.caller }

var hostname string

func init() {
	hostname, _ = os.Hostname()
}

// New returns a new Event.
// Event names are dot-separated for namespacing.
// Event identifiers identify the TTN entities that are related to the event.
// System events have nil identifiers.
// Event data will in most cases be marshaled to JSON, but ideally is a proto message.
func New(ctx context.Context, name string, identifiers CombinedIdentifiers, data interface{}) Event {
	evt := &event{
		ctx: ctx,
		innerEvent: ttnpb.Event{
			Name:           name,
			Time:           time.Now().UTC(),
			Origin:         hostname,
			CorrelationIDs: CorrelationIDsFromContext(ctx),
		},
		data: data,
	}
	if data, ok := data.(interface{ GetCorrelationIDs() []string }); ok {
		evt.innerEvent.CorrelationIDs = append(evt.innerEvent.CorrelationIDs, data.GetCorrelationIDs()...)
	}
	if identifiers != nil {
		evt.innerEvent.Identifiers = identifiers.CombinedIdentifiers().GetEntityIdentifiers()
	}
	return evt
}

// Proto returns the protobuf representation of the event.
func Proto(e Event) (*ttnpb.Event, error) {
	evt := local(e)
	pb := evt.innerEvent
	ctx, err := marshalContext(e.Context())
	if err != nil {
		return nil, err
	}
	pb.Context = ctx
	if evt.data != nil {
		var err error
		if protoMessage, ok := evt.data.(proto.Message); ok {
			pb.Data, err = types.MarshalAny(protoMessage)
		} else if errData, ok := evt.data.(error); ok {
			if ttnErrData, ok := errors.From(errData); ok {
				pb.Data, err = types.MarshalAny(ttnpb.ErrorDetailsToProto(ttnErrData))
			} else {
				pb.Data, err = types.MarshalAny(&types.StringValue{Value: errData.Error()})
			}
		} else {
			value, err := gogoproto.Value(evt.data)
			if err != nil {
				return nil, err
			}
			if _, isNull := value.Kind.(*types.Value_NullValue); !isNull {
				pb.Data, err = types.MarshalAny(value)
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return &pb, nil
}

// FromProto returns the event from its protobuf representation.
func FromProto(pb *ttnpb.Event) (Event, error) {
	ctx, err := unmarshalContext(context.Background(), pb.Context)
	if err != nil {
		return nil, err
	}
	evt := &event{
		ctx:        ctx,
		innerEvent: *pb,
	}
	if evt.innerEvent.Data != nil {
		any, err := types.EmptyAny(evt.innerEvent.Data)
		if err != nil {
			return nil, err
		}
		err = types.UnmarshalAny(evt.innerEvent.Data, any)
		if err != nil {
			return nil, err
		}
		evt.data = any
		if value, ok := evt.data.(*types.Value); ok {
			evt.data, err = gogoproto.Interface(value)
			if err != nil {
				return nil, err
			}
		}
		evt.innerEvent.Data = nil
	}
	return evt, nil
}

// UnmarshalJSON unmarshals an event as JSON.
func UnmarshalJSON(data []byte) (Event, error) {
	e := new(event)
	err := json.Unmarshal(data, e)
	if err != nil {
		return nil, err
	}
	return e, nil
}
