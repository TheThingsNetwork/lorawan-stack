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
	return s.ContextFunc()
}

// SendMsg calls s.SendMsgFunc.
func (s *MockStream) SendMsg(m interface{}) error {
	return s.SendMsgFunc(m)
}

// RecvMsg calls s.RecvMsgFunc.
func (s *MockStream) RecvMsg(m interface{}) error {
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
	return s.SetHeaderFunc(md)
}

// SendHeader calls s.SendHeaderFunc.
func (s *MockServerStream) SendHeader(md metadata.MD) error {
	return s.SendHeaderFunc(md)
}

// SetTrailer calls s.SetTrailerFunc.
func (s *MockServerStream) SetTrailer(md metadata.MD) {
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
	return s.HeaderFunc()
}

// Trailer calls s.TrailerFunc.
func (s *MockClientStream) Trailer() metadata.MD {
	return s.TrailerFunc()
}

// CloseSend calls s.CloseSendFunc.
func (s *MockClientStream) CloseSend() error {
	return s.CloseSendFunc()
}
