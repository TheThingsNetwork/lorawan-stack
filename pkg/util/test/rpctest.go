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

package test

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// MockStream is a mock grpc.Stream.
type MockStream struct {
	ContextFunc func() context.Context
	SendMsgFunc func(m interface{}) error
	RecvMsgFunc func(m interface{}) error
}

// Context calls s.ContextFunc.
func (s *MockStream) Context() context.Context {
	if s == nil || s.ContextFunc == nil {
		return nil
	}
	return s.ContextFunc()
}

// SendMsg calls s.SendMsgFunc.
func (s *MockStream) SendMsg(m interface{}) error {
	if s == nil || s.SendMsgFunc == nil {
		return nil
	}
	return s.SendMsgFunc(m)
}

// RecvMsg calls s.RecvMsgFunc.
func (s *MockStream) RecvMsg(m interface{}) error {
	if s == nil || s.RecvMsgFunc == nil {
		return nil
	}
	return s.RecvMsgFunc(m)
}

// MockServerStream is a mock grpc.ServerStream.
type MockServerStream struct {
	*MockStream
	SetHeaderFunc  func(md metadata.MD) error
	SendHeaderFunc func(md metadata.MD) error
	SetTrailerFunc func(md metadata.MD)
}

// SetHeader calls s.SetHeaderFunc.
func (s *MockServerStream) SetHeader(md metadata.MD) error {
	if s == nil || s.SetHeaderFunc == nil {
		return nil
	}
	return s.SetHeaderFunc(md)
}

// SendHeader calls s.SendHeaderFunc.
func (s *MockServerStream) SendHeader(md metadata.MD) error {
	if s == nil || s.SendHeaderFunc == nil {
		return nil
	}
	return s.SendHeaderFunc(md)
}

// SetTrailer calls s.SetTrailerFunc.
func (s *MockServerStream) SetTrailer(md metadata.MD) {
	if s == nil || s.SetTrailerFunc == nil {
		return
	}
	s.SetTrailerFunc(md)
}

// MockClientStream is a mock grpc.ClientStream.
type MockClientStream struct {
	*MockStream
	HeaderFunc    func() (metadata.MD, error)
	TrailerFunc   func() metadata.MD
	CloseSendFunc func() error
}

// Header calls s.HeaderFunc.
func (s *MockClientStream) Header() (metadata.MD, error) {
	if s == nil || s.HeaderFunc == nil {
		return metadata.MD{}, nil
	}
	return s.HeaderFunc()
}

// Trailer calls s.TrailerFunc.
func (s *MockClientStream) Trailer() metadata.MD {
	if s == nil || s.TrailerFunc == nil {
		return metadata.MD{}
	}
	return s.TrailerFunc()
}

// CloseSend calls s.CloseSendFunc.
func (s *MockClientStream) CloseSend() error {
	if s == nil || s.CloseSendFunc == nil {
		return nil
	}
	return s.CloseSendFunc()
}

// MockServerTransportStream is a mock grpc.ServerTransportStream.
type MockServerTransportStream struct {
	*MockServerStream
	MethodFunc     func() string
	SetTrailerFunc func(metadata.MD) error
}

// Method calls s.MethodFunc.
func (s *MockServerTransportStream) Method() string {
	if s == nil || s.MethodFunc == nil {
		return ""
	}
	return s.MethodFunc()
}

// Method calls s.SetTrailerFunc or s.MockServerStream.SetTrailer if s.SetTrailerFunc is nil.
func (s *MockServerTransportStream) SetTrailer(md metadata.MD) error {
	if s == nil {
		return nil
	}

	if s.SetTrailerFunc != nil {
		return s.SetTrailerFunc(md)
	}
	if s.MockServerStream != nil {
		s.MockServerStream.SetTrailer(md)
	}
	return nil
}
