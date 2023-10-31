// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package protocol_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/console/internal/events/protocol"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMarshal(t *testing.T) {
	t.Parallel()

	a := assertions.New(t)

	b, err := json.Marshal(protocol.MessageTypePublish)
	if a.So(err, should.BeNil) {
		a.So(b, should.Resemble, []byte(`"publish"`))
	}
	var tp protocol.MessageType
	err = json.Unmarshal([]byte(`"publish"`), &tp)
	if a.So(err, should.BeNil) {
		a.So(tp, should.Equal, protocol.MessageTypePublish)
	}

	b, err = json.Marshal(&protocol.SubscribeRequest{
		ID: 0x42,
		Identifiers: []*ttnpb.EntityIdentifiers{
			(&ttnpb.ApplicationIdentifiers{ApplicationId: "foo"}).GetEntityIdentifiers(),
			(&ttnpb.ClientIdentifiers{ClientId: "bar"}).GetEntityIdentifiers(),
		},
		Tail:  10,
		After: timePtr(time.UnixMilli(123456789012).UTC()),
		Names: []string{"foo", "bar"},
	})
	if a.So(err, should.BeNil) {
		a.So(
			b,
			should.Resemble,
			[]byte(`{"type":"subscribe","id":66,"identifiers":[{"application_ids":{"application_id":"foo"}},{"client_ids":{"client_id":"bar"}}],"tail":10,"after":"1973-11-29T21:33:09.012Z","names":["foo","bar"]}`), // nolint:lll
		)
	}
	var subReq protocol.SubscribeRequest
	err = json.Unmarshal(
		[]byte(`{"type":"subscribe","id":66,"identifiers":[{"application_ids":{"application_id":"foo"}},{"client_ids":{"client_id":"bar"}}],"tail":10,"after":"1973-11-29T21:33:09.012Z","names":["foo","bar"]}`), // nolint:lll
		&subReq,
	)
	if a.So(err, should.BeNil) {
		a.So(subReq, should.Resemble, protocol.SubscribeRequest{
			ID: 0x42,
			Identifiers: []*ttnpb.EntityIdentifiers{
				(&ttnpb.ApplicationIdentifiers{ApplicationId: "foo"}).GetEntityIdentifiers(),
				(&ttnpb.ClientIdentifiers{ClientId: "bar"}).GetEntityIdentifiers(),
			},
			Tail:  10,
			After: timePtr(time.UnixMilli(123456789012).UTC()),
			Names: []string{"foo", "bar"},
		})
	}

	b, err = json.Marshal(&protocol.SubscribeResponse{
		ID: 0x42,
	})
	if a.So(err, should.BeNil) {
		a.So(b, should.Resemble, []byte(`{"type":"subscribe","id":66}`))
	}
	var subResp protocol.SubscribeResponse
	err = json.Unmarshal([]byte(`{"type":"subscribe","id":66}`), &subResp)
	if a.So(err, should.BeNil) {
		a.So(subResp, should.Resemble, protocol.SubscribeResponse{ID: 0x42})
	}

	b, err = json.Marshal(&protocol.UnsubscribeRequest{
		ID: 0x42,
	})
	if a.So(err, should.BeNil) {
		a.So(b, should.Resemble, []byte(`{"type":"unsubscribe","id":66}`))
	}
	var unsubReq protocol.UnsubscribeRequest
	err = json.Unmarshal([]byte(`{"type":"unsubscribe","id":66}`), &unsubReq)
	if a.So(err, should.BeNil) {
		a.So(unsubReq, should.Resemble, protocol.UnsubscribeRequest{ID: 0x42})
	}

	b, err = json.Marshal(&protocol.UnsubscribeResponse{
		ID: 0x42,
	})
	if a.So(err, should.BeNil) {
		a.So(b, should.Resemble, []byte(`{"type":"unsubscribe","id":66}`))
	}
	var unsubResp protocol.UnsubscribeResponse
	err = json.Unmarshal([]byte(`{"type":"unsubscribe","id":66}`), &unsubResp)
	if a.So(err, should.BeNil) {
		a.So(unsubResp, should.Resemble, protocol.UnsubscribeResponse{ID: 0x42})
	}

	b, err = json.Marshal(&protocol.PublishResponse{
		ID: 0x42,
		Event: &ttnpb.Event{
			Name: "foo",
			Time: timestamppb.New(time.UnixMilli(123456789012).UTC()),
			Identifiers: []*ttnpb.EntityIdentifiers{
				(&ttnpb.ApplicationIdentifiers{ApplicationId: "foo"}).GetEntityIdentifiers(),
			},
			Data: test.Must(anypb.New(&ttnpb.ApplicationUp{
				Up: &ttnpb.ApplicationUp_UplinkMessage{
					UplinkMessage: &ttnpb.ApplicationUplink{},
				},
			})),
			CorrelationIds: []string{"foo", "bar"},
		},
	})
	if a.So(err, should.BeNil) {
		a.So(
			b,
			should.Resemble,
			[]byte(`{"type":"publish","id":66,"event":{"name":"foo","time":"1973-11-29T21:33:09.012Z","identifiers":[{"application_ids":{"application_id":"foo"}}],"data":{"@type":"type.googleapis.com/ttn.lorawan.v3.ApplicationUp","uplink_message":{}},"correlation_ids":["foo","bar"]}}`), // nolint:lll
		)
	}
	var pubResp protocol.PublishResponse
	err = json.Unmarshal(
		[]byte(`{"type":"publish","id":66,"event":{"name":"foo","time":"1973-11-29T21:33:09.012Z","identifiers":[{"application_ids":{"application_id":"foo"}}],"data":{"@type":"type.googleapis.com/ttn.lorawan.v3.ApplicationUp","uplink_message":{}},"correlation_ids":["foo","bar"]}}`), // nolint:lll
		&pubResp,
	)
	if a.So(err, should.BeNil) {
		a.So(pubResp, should.Resemble, protocol.PublishResponse{
			ID: 0x42,
			Event: &ttnpb.Event{
				Name: "foo",
				Time: timestamppb.New(time.UnixMilli(123456789012).UTC()),
				Identifiers: []*ttnpb.EntityIdentifiers{
					(&ttnpb.ApplicationIdentifiers{ApplicationId: "foo"}).GetEntityIdentifiers(),
				},
				Data: test.Must(anypb.New(&ttnpb.ApplicationUp{
					Up: &ttnpb.ApplicationUp_UplinkMessage{
						UplinkMessage: &ttnpb.ApplicationUplink{},
					},
				})),
				CorrelationIds: []string{"foo", "bar"},
			},
		})
	}

	errDefinition := errors.DefineInvalidArgument("bad_argument", "bad argument `{argument}`")
	errInstance := errDefinition.WithAttributes("argument", "foo")
	errStatus := status.Convert(errInstance)
	errJSON := test.Must(json.Marshal(errInstance))
	b, err = json.Marshal(&protocol.ErrorResponse{
		ID:    0x42,
		Error: errStatus,
	})
	if a.So(err, should.BeNil) {
		a.So(b, should.Resemble, []byte(fmt.Sprintf(`{"type":"error","id":66,"error":%v}`, string(errJSON)))) // nolint:lll
	}
	var errResp protocol.ErrorResponse
	err = json.Unmarshal([]byte(fmt.Sprintf(`{"type":"error","id":66,"error":%v}`, string(errJSON))), &errResp) // nolint:lll
	if a.So(err, should.BeNil) {
		a.So(errResp, should.Resemble, protocol.ErrorResponse{
			ID:    0x42,
			Error: errStatus,
		})
	}

	var reqWrapper protocol.RequestWrapper
	err = json.Unmarshal(
		[]byte(`{"type":"subscribe","id":66,"identifiers":[{"application_ids":{"application_id":"foo"}},{"client_ids":{"client_id":"bar"}}],"tail":10,"after":"1973-11-29T21:33:09.012Z","names":["foo","bar"]}`), // nolint:lll
		&reqWrapper,
	)
	if a.So(err, should.BeNil) {
		a.So(reqWrapper, should.Resemble, protocol.RequestWrapper{
			Contents: &protocol.SubscribeRequest{
				ID: 0x42,
				Identifiers: []*ttnpb.EntityIdentifiers{
					(&ttnpb.ApplicationIdentifiers{ApplicationId: "foo"}).GetEntityIdentifiers(),
					(&ttnpb.ClientIdentifiers{ClientId: "bar"}).GetEntityIdentifiers(),
				},
				Tail:  10,
				After: timePtr(time.UnixMilli(123456789012).UTC()),
				Names: []string{"foo", "bar"},
			},
		})
	}

	var respWrapper protocol.ResponseWrapper
	err = json.Unmarshal([]byte(`{"type":"subscribe","id":66}`), &respWrapper)
	if a.So(err, should.BeNil) {
		a.So(respWrapper, should.Resemble, protocol.ResponseWrapper{
			Contents: &protocol.SubscribeResponse{
				ID: 0x42,
			},
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
