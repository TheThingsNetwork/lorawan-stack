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

// Context calls ContextFunc if set and panics otherwise.
func (m MockStream) Context() context.Context {
	if m.ContextFunc == nil {
		panic("Context called, but not set")
	}
	return m.ContextFunc()
}

// SendMsg calls SendMsgFunc if set and panics otherwise.
func (m MockStream) SendMsg(msg interface{}) error {
	if m.SendMsgFunc == nil {
		panic("SendMsg called, but not set")
	}
	return m.SendMsgFunc(msg)
}

// RecvMsg calls RecvMsgFunc if set and panics otherwise.
func (m MockStream) RecvMsg(msg interface{}) error {
	if m.RecvMsgFunc == nil {
		panic("RecvMsg called, but not set")
	}
	return m.RecvMsgFunc(msg)
}

// MockServerStream is a mock grpc.ServerStream.
type MockServerStream struct {
	*MockStream
	SetHeaderFunc  func(md metadata.MD) error
	SendHeaderFunc func(md metadata.MD) error
	SetTrailerFunc func(md metadata.MD)
}

// SetHeader calls SetHeaderFunc if set and panics otherwise.
func (m MockServerStream) SetHeader(md metadata.MD) error {
	if m.SetHeaderFunc == nil {
		panic("SetHeader called, but not set")
	}
	return m.SetHeaderFunc(md)
}

// SendHeader calls SendHeaderFunc if set and panics otherwise.
func (m MockServerStream) SendHeader(md metadata.MD) error {
	if m.SendHeaderFunc == nil {
		panic("SendHeader called, but not set")
	}
	return m.SendHeaderFunc(md)
}

// SetTrailer calls SetTrailerFunc if set and panics otherwise.
func (m MockServerStream) SetTrailer(md metadata.MD) {
	if m.SetTrailerFunc == nil {
		panic("X called, but not set")
	}
	m.SetTrailerFunc(md)
}

// MockClientStream is a mock grpc.ClientStream.
type MockClientStream struct {
	*MockStream
	HeaderFunc    func() (metadata.MD, error)
	TrailerFunc   func() metadata.MD
	CloseSendFunc func() error
}

// Header calls HeaderFunc if set and panics otherwise.
func (m MockClientStream) Header() (metadata.MD, error) {
	if m.HeaderFunc == nil {
		panic("Header called, but not set")
	}
	return m.HeaderFunc()
}

// Trailer calls TrailerFunc if set and panics otherwise.
func (m MockClientStream) Trailer() metadata.MD {
	if m.TrailerFunc == nil {
		panic("Trailer called, but not set")
	}
	return m.TrailerFunc()
}

// CloseSend calls CloseSendFunc if set and panics otherwise.
func (m MockClientStream) CloseSend() error {
	if m.CloseSendFunc == nil {
		panic("CloseSend called, but not set")
	}
	return m.CloseSendFunc()
}

// MockServerTransportStream is a mock grpc.ServerTransportStream.
type MockServerTransportStream struct {
	*MockServerStream
	MethodFunc     func() string
	SetTrailerFunc func(metadata.MD) error
}

// Method calls MethodFunc if set and panics otherwise.
func (m MockServerTransportStream) Method() string {
	if m.MethodFunc == nil {
		panic("Method called, but not set")
	}
	return m.MethodFunc()
}

// SetTrailer calls SetTrailerFunc if set and panics otherwise.
func (m MockServerTransportStream) SetTrailer(md metadata.MD) error {
	if m.SetTrailerFunc == nil {
		panic("SetTrailer called, but not set")
	}
	return m.SetTrailerFunc(md)
}
