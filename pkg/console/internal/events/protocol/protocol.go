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

// Package protocol implements the protocol for the events package.
package protocol

import (
	"encoding/json"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/status"
)

var (
	errMessageType = errors.DefineInvalidArgument("message_type", "invalid message type `{type}`")

	_ json.Marshaler   = (*ttnpb.EntityIdentifiers)(nil)
	_ json.Unmarshaler = (*ttnpb.EntityIdentifiers)(nil)

	_ json.Marshaler   = (*ttnpb.Event)(nil)
	_ json.Unmarshaler = (*ttnpb.Event)(nil)
)

// MessageType is the type of a message.
type MessageType int

const (
	// MessageTypeSubscribe is the type of a subscribe message.
	MessageTypeSubscribe MessageType = iota
	// MessageTypeUnsubscribe is the type of an unsubscribe message.
	MessageTypeUnsubscribe
	// MessageTypePublish is the type of a publish message.
	MessageTypePublish
	// MessageTypeError is the type of an error message.
	MessageTypeError
)

// MarshalJSON implements json.Marshaler.
func (m MessageType) MarshalJSON() ([]byte, error) {
	switch m {
	case MessageTypeSubscribe:
		return []byte(`"subscribe"`), nil
	case MessageTypeUnsubscribe:
		return []byte(`"unsubscribe"`), nil
	case MessageTypePublish:
		return []byte(`"publish"`), nil
	case MessageTypeError:
		return []byte(`"error"`), nil
	default:
		return nil, errMessageType.WithAttributes("type", m)
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *MessageType) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case `"subscribe"`:
		*m = MessageTypeSubscribe
	case `"unsubscribe"`:
		*m = MessageTypeUnsubscribe
	case `"publish"`:
		*m = MessageTypePublish
	case `"error"`:
		*m = MessageTypeError
	default:
		return errMessageType.WithAttributes("type", string(data))
	}
	return nil
}

// Request is a request message.
type Request interface {
	_requestMessage()
}

// Response is a response message.
type Response interface {
	_responseMessage()
}

// SubscribeRequest is the request to subscribe to events.
type SubscribeRequest struct {
	ID          uint64                     `json:"id"`
	Identifiers []*ttnpb.EntityIdentifiers `json:"identifiers"`
	Tail        uint32                     `json:"tail"`
	After       *time.Time                 `json:"after"`
	Names       []string                   `json:"names"`
}

func (SubscribeRequest) _requestMessage() {}

// Response builds a response to the request.
func (m SubscribeRequest) Response(err error) Response {
	if err != nil {
		return newErrorResponse(m.ID, err)
	}
	return &SubscribeResponse{
		ID: m.ID,
	}
}

// MarshalJSON implements json.Marshaler.
func (m SubscribeRequest) MarshalJSON() ([]byte, error) {
	type alias SubscribeRequest
	return jsonpb.TTN().Marshal(struct {
		Type MessageType `json:"type"`
		alias
	}{
		Type:  MessageTypeSubscribe,
		alias: alias(m),
	})
}

// SubscribeResponse is the response to a subscribe request.
type SubscribeResponse struct {
	ID uint64 `json:"id"`
}

func (SubscribeResponse) _responseMessage() {}

// MarshalJSON implements json.Marshaler.
func (m SubscribeResponse) MarshalJSON() ([]byte, error) {
	type alias SubscribeResponse
	return jsonpb.TTN().Marshal(struct {
		Type MessageType `json:"type"`
		alias
	}{
		Type:  MessageTypeSubscribe,
		alias: alias(m),
	})
}

// UnsubscribeRequest is the request to unsubscribe from events.
type UnsubscribeRequest struct {
	ID uint64 `json:"id"`
}

func (UnsubscribeRequest) _requestMessage() {}

// MarshalJSON implements json.Marshaler.
func (m UnsubscribeRequest) MarshalJSON() ([]byte, error) {
	type alias UnsubscribeRequest
	return jsonpb.TTN().Marshal(struct {
		Type MessageType `json:"type"`
		alias
	}{
		Type:  MessageTypeUnsubscribe,
		alias: alias(m),
	})
}

// UnsubscribeResponse is the response to an unsubscribe request.
type UnsubscribeResponse struct {
	ID uint64 `json:"id"`
}

func (UnsubscribeResponse) _responseMessage() {}

// Response builds a response to the request.
func (m UnsubscribeRequest) Response(err error) Response {
	if err != nil {
		return newErrorResponse(m.ID, err)
	}
	return &UnsubscribeResponse{
		ID: m.ID,
	}
}

// MarshalJSON implements json.Marshaler.
func (m UnsubscribeResponse) MarshalJSON() ([]byte, error) {
	type alias UnsubscribeResponse
	return jsonpb.TTN().Marshal(struct {
		Type MessageType `json:"type"`
		alias
	}{
		Type:  MessageTypeUnsubscribe,
		alias: alias(m),
	})
}

// PublishResponse is the request to publish an event.
type PublishResponse struct {
	ID    uint64       `json:"id"`
	Event *ttnpb.Event `json:"event"`
}

func (PublishResponse) _responseMessage() {}

// MarshalJSON implements json.Marshaler.
func (m PublishResponse) MarshalJSON() ([]byte, error) {
	type alias PublishResponse
	return jsonpb.TTN().Marshal(struct {
		Type MessageType `json:"type"`
		alias
	}{
		Type:  MessageTypePublish,
		alias: alias(m),
	})
}

// ErrorResponse is the response to an error.
type ErrorResponse struct {
	ID    uint64
	Error *status.Status
}

func (ErrorResponse) _responseMessage() {}

// statusAlias is an alias of status.Status which supports JSON marshaling.
type statusAlias statuspb.Status

// MarshalJSON implements json.Marshaler.
func (s *statusAlias) MarshalJSON() ([]byte, error) {
	return jsonpb.TTN().Marshal((*statuspb.Status)(s))
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *statusAlias) UnmarshalJSON(data []byte) error {
	return jsonpb.TTN().Unmarshal(data, (*statuspb.Status)(s))
}

// MarshalJSON implements json.Marshaler.
func (m ErrorResponse) MarshalJSON() ([]byte, error) {
	return jsonpb.TTN().Marshal(struct {
		Type  MessageType  `json:"type"`
		ID    uint64       `json:"id"`
		Error *statusAlias `json:"error"`
	}{
		Type:  MessageTypeError,
		ID:    m.ID,
		Error: (*statusAlias)(m.Error.Proto()),
	})
}

func newErrorResponse(id uint64, err error) Response {
	return &ErrorResponse{
		ID:    id,
		Error: status.Convert(err),
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *ErrorResponse) UnmarshalJSON(data []byte) error {
	var alias struct {
		ID    uint64       `json:"id"`
		Error *statusAlias `json:"error"`
	}
	if err := jsonpb.TTN().Unmarshal(data, &alias); err != nil {
		return err
	}
	m.ID = alias.ID
	m.Error = status.FromProto((*statuspb.Status)(alias.Error))
	return nil
}

// RequestWrapper wraps a request to be sent over the websocket.
type RequestWrapper struct {
	Contents Request
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *RequestWrapper) UnmarshalJSON(data []byte) error {
	var contents struct {
		Type MessageType `json:"type"`
	}
	if err := jsonpb.TTN().Unmarshal(data, &contents); err != nil {
		return err
	}
	switch contents.Type {
	case MessageTypeSubscribe:
		m.Contents = &SubscribeRequest{}
	case MessageTypeUnsubscribe:
		m.Contents = &UnsubscribeRequest{}
	default:
		return errMessageType.WithAttributes("type", contents.Type)
	}
	return jsonpb.TTN().Unmarshal(data, m.Contents)
}

// MarshalJSON implements json.Marshaler.
func (m RequestWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Contents)
}

// ResponseWrapper wraps a response to be sent over the websocket.
type ResponseWrapper struct {
	Contents Response
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *ResponseWrapper) UnmarshalJSON(data []byte) error {
	var contents struct {
		Type MessageType `json:"type"`
	}
	if err := jsonpb.TTN().Unmarshal(data, &contents); err != nil {
		return err
	}
	switch contents.Type {
	case MessageTypeSubscribe:
		m.Contents = &SubscribeResponse{}
	case MessageTypeUnsubscribe:
		m.Contents = &UnsubscribeResponse{}
	case MessageTypePublish:
		m.Contents = &PublishResponse{}
	case MessageTypeError:
		m.Contents = &ErrorResponse{}
	default:
		return errMessageType.WithAttributes("type", contents.Type)
	}
	return jsonpb.TTN().Unmarshal(data, m.Contents)
}

// MarshalJSON implements json.Marshaler.
func (m ResponseWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Contents)
}
