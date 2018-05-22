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
	"go.thethings.network/lorawan-stack/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FromGRPCStatus converts the gRPC status message into an Error.
func FromGRPCStatus(status *status.Status) Error {
	err := build(Definition{
		code:          int32(status.Code()),
		messageFormat: status.Message(),
	}, 0)
	for _, details := range status.Details() {
		switch details := details.(type) {
		case *ttnpb.ErrorDetails:
			err.namespace = details.Namespace
			err.name = details.Name
			err.messageFormat = details.MessageFormat
			err.correlationID = details.CorrelationID
			attributes, aErr := gogoproto.Map(details.Attributes)
			if aErr != nil {
				// TODO: this should probably panic
			} else {
				err.attributes = attributes
			}
		default:
			err.details = append(err.details, details)
		}
	}
	return err
}

// GRPCStatus converts the Definition into a gRPC status message.
func (d Definition) GRPCStatus() *status.Status {
	s := status.New(codes.Code(d.Code()), d.String())
	s, err := s.WithDetails(&ttnpb.ErrorDetails{
		Namespace:     d.namespace,
		Name:          d.name,
		MessageFormat: d.messageFormat,
	})
	if err != nil {
		// TODO: this should probably panic
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
	attributes, err := gogoproto.Struct(e.attributes)
	if err != nil {
		// TODO: this should probably panic
	}
	protoDetails = append(protoDetails, &ttnpb.ErrorDetails{
		Namespace:     e.namespace,
		Name:          e.name,
		MessageFormat: e.messageFormat,
		Attributes:    attributes,
		CorrelationID: e.correlationID,
	})
	s, err = s.WithDetails(protoDetails...)
	if err != nil {
		// TODO: this should probably panic
	}
	return s
}

// MarshalJSON implements json.Marshaler.
func (e Error) MarshalJSON() ([]byte, error) {
	return jsonpb.TTN().Marshal(e.GRPCStatus().Proto())
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *Error) UnmarshalJSON(data []byte) error {
	s := new(spb.Status)
	if err := jsonpb.TTN().Unmarshal(data, s); err != nil {
		return err
	}
	*e = FromGRPCStatus(status.FromProto(s))
	return nil
}
