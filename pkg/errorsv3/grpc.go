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

package errors

import (
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorDetails that can be carried over API.
type ErrorDetails interface {
	Namespace() string
	Name() string
	MessageFormat() string
	Attributes() map[string]interface{}
	CorrelationID() string
}

// FromGRPCStatus converts the gRPC status message into an Error.
func FromGRPCStatus(status *status.Status) Error {
	err := build(Definition{
		code:          int32(status.Code()),
		messageFormat: status.Message(),
	}, 0)
	if ErrorDetailsFromProto != nil {
		detailMsgs := status.Details()
		detailProtos := make([]proto.Message, 0, len(detailMsgs))
		for _, msg := range detailMsgs { // convert to []proto.Message
			if msg, ok := msg.(proto.Message); ok {
				detailProtos = append(detailProtos, msg)
			}
		}
		details, rest := ErrorDetailsFromProto(detailProtos...)
		if len(rest) != 0 {
			detailIfaces := make([]interface{}, len(rest))
			for i, iface := range rest { // convert to []interface{}
				detailIfaces[i] = iface
			}
			err.details = detailIfaces
		}
		if details != nil {
			if namespace := details.Namespace(); namespace != "" {
				err.namespace = namespace
			}
			if name := details.Name(); name != "" {
				err.name = name
			}
			if messageFormat := details.MessageFormat(); messageFormat != "" {
				err.messageFormat = messageFormat
			}
			if attributes := details.Attributes(); len(attributes) != 0 {
				err.attributes = attributes
			}
			if correlationID := details.CorrelationID(); correlationID != "" {
				err.correlationID = correlationID
			}
		}
	}
	return err
}

// ErrorDetailsToProto converts the given ErrorDetails into a protobuf-encoded message.
var ErrorDetailsToProto func(e ErrorDetails) (msg proto.Message)

// ErrorDetailsFromProto ranges over the given protobuf-encoded messages to extract the ErrorDetails. It returns any
var ErrorDetailsFromProto func(msg ...proto.Message) (details ErrorDetails, rest []proto.Message)

// GRPCStatus converts the Definition into a gRPC status message.
func (d Definition) GRPCStatus() *status.Status {
	s := status.New(codes.Code(d.Code()), d.String())
	if ErrorDetailsToProto != nil {
		if proto := ErrorDetailsToProto(d); proto != nil {
			var err error
			s, err = s.WithDetails(proto)
			if err != nil {
				// TODO: this should probably panic
			}
		}
	}
	return s
}

// GRPCStatus converts the Error into a gRPC status message.
func (e Error) GRPCStatus() *status.Status {
	s := status.New(codes.Code(e.Code()), e.String())
	protoDetails := make([]proto.Message, 0, len(e.Details())+1)
	for _, details := range e.Details() {
		if details, ok := details.(proto.Message); ok {
			protoDetails = append(protoDetails, details)
		}
	}
	if ErrorDetailsToProto != nil {
		if proto := ErrorDetailsToProto(e); proto != nil {
			protoDetails = append(protoDetails, proto)
		}
	}
	if len(protoDetails) != 0 {
		var err error
		s, err = s.WithDetails(protoDetails...)
		if err != nil {
			// TODO: this should probably panic
		}
	}
	return s
}
