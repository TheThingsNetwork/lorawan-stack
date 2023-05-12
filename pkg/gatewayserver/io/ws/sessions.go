// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package ws

import (
	"context"
	"sync"
)

type sessionKeyType struct{}

var sessionKey sessionKeyType

// Session contains the session state for a single gateway.
type Session struct {
	DataMu sync.RWMutex
	Data   any
}

// NewContextWithSession returns a new context with the session.
func NewContextWithSession(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionKey, session)
}

// SessionFromContext returns a new session from the context.
// The session value can be modified by the caller.
func SessionFromContext(ctx context.Context) *Session {
	if session, ok := ctx.Value(sessionKey).(*Session); ok {
		return session
	}
	return nil
}
