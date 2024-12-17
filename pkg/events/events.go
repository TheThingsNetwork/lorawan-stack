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
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	sentryerrors "go.thethings.network/lorawan-stack/v3/pkg/errors/sentry"
	"go.thethings.network/lorawan-stack/v3/pkg/goproto"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Event interface.
type Event interface {
	UniqueID() string
	Context() context.Context
	Name() string
	Time() time.Time
	Identifiers() []*ttnpb.EntityIdentifiers
	Data() any
	CorrelationIds() []string
	Origin() string
	Caller() string
	Visibility() *ttnpb.Rights
	AuthType() string
	AuthTokenID() string
	AuthTokenType() string
	RemoteIP() string
	UserAgent() string
}

func local(evt Event) *event {
	localEvent, ok := evt.(*event)
	if !ok {
		t := evt.Time()
		localEvent = &event{
			ctx: evt.Context(),
			innerEvent: &ttnpb.Event{
				UniqueId:       evt.UniqueID(),
				Name:           evt.Name(),
				Time:           timestamppb.New(t),
				Identifiers:    evt.Identifiers(),
				CorrelationIds: evt.CorrelationIds(),
				Origin:         evt.Origin(),
				Visibility:     evt.Visibility(),
				UserAgent:      evt.UserAgent(),
				RemoteIp:       evt.RemoteIP(),
			},
			data:   evt.Data(),
			caller: evt.Caller(),
		}
		authentication := &ttnpb.Event_Authentication{
			Type:      evt.AuthType(),
			TokenType: evt.AuthTokenType(),
			TokenId:   evt.AuthTokenID(),
		}
		if authentication.TokenId != "" || authentication.TokenType != "" || authentication.Type != "" {
			localEvent.innerEvent.Authentication = authentication
		}
	}
	return localEvent
}

type event struct {
	ctx        context.Context
	innerEvent *ttnpb.Event
	data       any
	caller     string
}

var pathPrefix = func() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("could not determine location of events.go")
	}
	return strings.TrimSuffix(file, filepath.Join("pkg", "events", "events.go"))
}()

// IncludeCaller indicates whether the caller of Publish should be included in the event.
var IncludeCaller bool

// withCaller returns an event with the Caller field populated, if configured to do so.
// If the original event already had a non-empty Caller, the original event is returned.
func (e *event) withCaller() *event {
	if IncludeCaller && e.caller == "" {
		if _, file, line, ok := runtime.Caller(2); ok {
			clone := *e
			clone.caller = fmt.Sprintf("%s:%d", strings.TrimPrefix(file, pathPrefix), line)
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

func (e event) UniqueID() string         { return e.innerEvent.UniqueId }
func (e event) Context() context.Context { return e.ctx }
func (e event) Name() string             { return e.innerEvent.Name }
func (e event) Time() time.Time {
	t := ttnpb.StdTime(e.innerEvent.GetTime())
	if t != nil {
		return *t
	}
	return time.Time{}
}
func (e event) Identifiers() []*ttnpb.EntityIdentifiers { return e.innerEvent.Identifiers }
func (e event) Data() any                               { return e.data }
func (e event) CorrelationIds() []string                { return e.innerEvent.CorrelationIds }
func (e event) Origin() string                          { return e.innerEvent.Origin }
func (e event) Caller() string                          { return e.caller }
func (e event) Visibility() *ttnpb.Rights               { return e.innerEvent.Visibility }
func (e event) UserAgent() string                       { return e.innerEvent.UserAgent }
func (e event) RemoteIP() string                        { return e.innerEvent.RemoteIp }
func (e event) AuthType() string                        { return e.innerEvent.GetAuthentication().GetType() }
func (e event) AuthTokenType() string                   { return e.innerEvent.GetAuthentication().GetTokenType() }
func (e event) AuthTokenID() string                     { return e.innerEvent.GetAuthentication().GetTokenId() }

var hostname string

func init() {
	hostname, _ = os.Hostname()
}

// New returns a new Event.
// Instead of using New, most implementations should first define an event,
// and then create a new event from that definition.
func New(ctx context.Context, name, description string, opts ...Option) Event {
	return (&definition{name: name, description: description}).New(ctx, opts...)
}

var errMarshalData = errors.Define("marshal_data", "marshal data")

func marshalData(data any) (anyPB *anypb.Any, err error) {
	// TODO: https://github.com/TheThingsIndustries/lorawan-stack-support/issues/1163.
	// Remove this after the issue is fixed.
	defer func() {
		if p := recover(); p != nil {
			if pErr, ok := p.(error); ok {
				err = errMarshalData.WithCause(pErr)
			} else {
				err = errMarshalData.WithAttributes("panic", p)
			}
			event := sentryerrors.NewEvent(err)
			sentry.CaptureEvent(event)
		}
	}()

	return mustMarshalData(data)
}

func mustMarshalData(data any) (anyPB *anypb.Any, err error) {
	if protoMessage, ok := data.(proto.Message); ok {
		anyPB, err = anypb.New(protoMessage)
		if err != nil {
			return nil, err
		}
	} else if errData, ok := data.(error); ok {
		if ttnErrData, ok := errors.From(errData); ok {
			anyPB, err = anypb.New(ttnpb.ErrorDetailsToProto(ttnErrData))
			if err != nil {
				return nil, err
			}
		} else {
			anyPB, err = anypb.New(&wrapperspb.StringValue{Value: errData.Error()})
			if err != nil {
				return nil, err
			}
		}
	} else {
		value, err := goproto.Value(data)
		if err != nil {
			return nil, err
		}
		if _, isNull := value.Kind.(*structpb.Value_NullValue); !isNull {
			anyPB, err = anypb.New(value)
			if err != nil {
				return nil, err
			}
		}
	}
	return anyPB, nil
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
		pb.Data, err = marshalData(e.Data())
		if err != nil {
			return nil, err
		}
	}
	return pb, nil
}

// FromProto returns the event from its protobuf representation.
func FromProto(pb *ttnpb.Event) (Event, error) {
	ctx, err := unmarshalContext(context.Background(), pb.Context)
	if err != nil {
		return nil, err
	}
	var data any
	if pb.Data != nil {
		anyMsg, err := pb.Data.UnmarshalNew()
		if err != nil {
			return nil, err
		}
		data = anyMsg
		v, ok := anyMsg.(*structpb.Value)
		if ok {
			iface, err := goproto.Interface(v)
			if err != nil {
				return nil, err
			}
			data = iface
		}
	}
	return &event{
		ctx:  ctx,
		data: data,
		innerEvent: &ttnpb.Event{
			UniqueId:       pb.UniqueId,
			Name:           pb.Name,
			Time:           pb.Time,
			Identifiers:    pb.Identifiers,
			CorrelationIds: pb.CorrelationIds,
			Origin:         pb.Origin,
			Visibility:     pb.Visibility,
			Authentication: pb.Authentication,
			RemoteIp:       pb.RemoteIp,
			UserAgent:      pb.UserAgent,
		},
	}, nil
}

// UnmarshalJSON unmarshals an event as JSON.
func UnmarshalJSON(data []byte) (Event, error) {
	e := new(event)
	if err := json.Unmarshal(data, e); err != nil {
		return nil, err
	}
	return e, nil
}
