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

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

// Session contains the session state for a single gateway.
type Session struct {
	ID int32
}

var (
	errSessionNotFound      = errors.DefineNotFound("session_not_found", "session not found")
	errSessionAlreadyExists = errors.DefineAlreadyExists("session_already_exists", "session already exists")
)

// Sessions holds state of the WS sessions.
type Sessions struct {
	items sync.Map
}

// NewSession creates a new session for the given UID.
// If errSessionAlreadyExists is returned, it indicates a corrupted state or improper session termination.
func (s *Sessions) NewSession(ctx context.Context, uid string) error {
	if _, ok := s.items.Load(uid); ok {
		return errSessionAlreadyExists
	}
	s.items.Store(uid, Session{})
	return nil
}

// GetSession retrieves the session for the given UID.
func (s *Sessions) GetSession(uid string) (Session, error) {
	val, ok := s.items.Load(uid)
	if !ok {
		return Session{}, errSessionNotFound
	}
	return val.(Session), nil
}

// UpdateSession updates the session state for the given UID.
// If errSessionNotFound is returned, it indicates a corrupted state or improper session termination.
func (s *Sessions) UpdateSession(uid string, session Session) error {
	_, ok := s.items.Load(uid)
	if !ok {
		return errSessionNotFound
	}
	s.items.Store(uid, session)
	return nil
}

// DeleteSession session removes the session for the UID.
// This function must be called at termination of a gateway connection.
// If errSessionNotFound is returned, it indicates a corrupted state or improper session termination.
func (s *Sessions) DeleteSession(uid string) error {
	_, ok := s.items.Load(uid)
	if !ok {
		return errSessionNotFound
	}
	s.items.Delete(uid)
	return nil
}
