// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

// Package mock provides a mock sink for testing.
package mock

import (
	"net/http"
	"sync/atomic"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/sink"
)

// Sink is a sink that can be used for testing.
type Sink interface {
	sink.Sink

	SetError(error)
}

type mockSink struct {
	ch  chan<- *http.Request
	err atomic.Pointer[error]
}

// Process implements Sink.
func (s *mockSink) Process(req *http.Request) (err error) {
	select {
	case <-req.Context().Done():
		return req.Context().Err()
	case s.ch <- req:
		if pErr := s.err.Load(); pErr != nil {
			return *pErr
		}
		return nil
	}
}

// SetError sets the error to be returned by Process.
func (s *mockSink) SetError(err error) {
	if err != nil {
		s.err.Store(&err)
	} else {
		s.err.Store(nil)
	}
}

// New returns a new channel sink.
func New(ch chan<- *http.Request) Sink {
	return &mockSink{
		ch: ch,
	}
}
