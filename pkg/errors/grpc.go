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

package errors

import (
	"context"
	"encoding/hex"

	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FromGRPCStatus converts the gRPC status message into an Error.
func FromGRPCStatus(status *status.Status) Error {
	err := build(Definition{
		code:          uint32(status.Code()),
		messageFormat: status.Message(),
	}, 0)
	if ErrorDetailsFromProto == nil {
		return err
	}
	detailMsgs := status.Details()
	detailProtos := make([]proto.Message, 0, len(detailMsgs))
	for _, msg := range detailMsgs { // convert to []proto.Message
		if msg, ok := msg.(proto.Message); ok {
			detailProtos = append(detailProtos, msg)
		}
	}
	details, rest := ErrorDetailsFromProto(detailProtos...)
	if len(rest) != 0 {
		err.details = rest
	}
	if details != nil {
		setErrorDetails(&err, details)
	}
	return err
}

// ErrorDetailsToProto converts the given ErrorDetails into a protobuf-encoded message.
//
// This variable is set by pkg/ttnpb.
var ErrorDetailsToProto func(e ErrorDetails) (msg proto.Message)

// ErrorDetailsFromProto ranges over the given protobuf-encoded messages
// to extract the ErrorDetails. It returns details if present, as well as the
// rest of the details.
//
// This variable is set by pkg/ttnpb.
var ErrorDetailsFromProto func(msg ...proto.Message) (details ErrorDetails, rest []proto.Message)

// setGRPCStatus sets a (marshaled) gRPC status in the error definition.
//
// This func should be called when the error definition is created. Doing that
// makes that we have to convert to a gRPC status only once instead of on every call.
func (d *Definition) setGRPCStatus() {
	s := status.New(codes.Code(d.Code()), d.String())
	if ErrorDetailsToProto != nil {
		if pb := ErrorDetailsToProto(d); pb != nil {
			var err error
			s, err = s.WithDetails(pb)
			if err != nil {
				panic(err) // ErrorDetailsToProto generated an invalid proto.
			}
		}
	}
	d.grpcStatus = s
}

// GRPCStatus returns the Definition as a gRPC status message.
func (d Definition) GRPCStatus() *status.Status {
	return d.grpcStatus // initialized when defined (with setGRPCStatus).
}

func (e *Error) clearGRPCStatus() {
	e.grpcStatus.Store((*status.Status)(nil))
}

// GRPCStatus converts the Error into a gRPC status message.
func (e *Error) GRPCStatus() *status.Status {
	if s, ok := e.grpcStatus.Load().(*status.Status); ok && s != nil {
		return s
	}
	s := status.New(codes.Code(e.Code()), e.String())
	if ErrorDetailsToProto != nil {
		if pb := ErrorDetailsToProto(e); pb != nil {
			var err error
			s, err = s.WithDetails(pb)
			if err != nil {
				panic(err) // ErrorDetailsToProto generated an invalid proto.
			}
		}
	}
	e.grpcStatus.Store(s)
	return s
}

// UnaryServerInterceptor makes sure that returned TTN errors contain a CorrelationID.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		res, err := handler(ctx, req)
		if ttnErr, ok := From(err); ok {
			if ttnErr.correlationID == "" {
				ttnErr.correlationID = hex.EncodeToString(uuid.NewV4().Bytes()) // Compliant with Sentry.
			}
			err = ttnErr
		}
		return res, err
	}
}

// StreamServerInterceptor makes sure that returned TTN errors contain a CorrelationID.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, stream)
		if ttnErr, ok := From(err); ok {
			if ttnErr.correlationID == "" {
				ttnErr.correlationID = hex.EncodeToString(uuid.NewV4().Bytes()) // Compliant with Sentry.
			}
			err = ttnErr
		}
		return err
	}
}

// UnaryClientInterceptor converts gRPC errors to regular errors.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if ttnErr, ok := From(err); ok {
			return ttnErr
		}
		return err
	}
}

type wrappedStream struct {
	grpc.ClientStream
}

func (w wrappedStream) SendMsg(m interface{}) error {
	err := w.ClientStream.SendMsg(m)
	if ttnErr, ok := From(err); ok {
		return ttnErr
	}
	return err
}
func (w wrappedStream) RecvMsg(m interface{}) error {
	err := w.ClientStream.RecvMsg(m)
	if ttnErr, ok := From(err); ok {
		return ttnErr
	}
	return err
}

// StreamClientInterceptor converts gRPC errors to regular errors.
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		s, err := streamer(ctx, desc, cc, method, opts...)
		if ttnErr, ok := From(err); ok {
			return nil, ttnErr
		}
		if err != nil {
			return nil, err
		}
		return wrappedStream{s}, nil
	}
}
