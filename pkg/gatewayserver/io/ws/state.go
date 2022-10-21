// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
)

// state represents the LBS session state.
type state struct {
	ID       *int32
	TimeSync *bool
}

// updateState updates the session state.
func updateState(ctx context.Context, f func(*state)) {
	session := SessionFromContext(ctx)
	session.DataMu.Lock()
	defer session.DataMu.Unlock()
	st, ok := session.Data.(*state)
	if !ok {
		st = &state{}
		session.Data = st
	}
	f(st)
}

// GetState returns the session state.
func getState(ctx context.Context, f func(*state) interface{}) interface{} {
	session := SessionFromContext(ctx)
	session.DataMu.RLock()
	defer session.DataMu.RUnlock()
	st, ok := session.Data.(*state)
	if !ok {
		return nil
	}
	return f(st)
}

// UpdateSessionID updates the session ID.
func UpdateSessionID(ctx context.Context, id int32) {
	updateState(ctx, func(st *state) {
		st.ID = &id
	})
}

// UpdateSessionTimeSync updates the session time sync.
func UpdateSessionTimeSync(ctx context.Context, b bool) {
	updateState(ctx, func(st *state) {
		st.TimeSync = &b
	})
}

// GetSessionID returns the session ID.
func GetSessionID(ctx context.Context) (int32, bool) {
	i, ok := getState(ctx, func(st *state) interface{} {
		if st.ID != nil {
			return *st.ID
		}
		return nil
	}).(int32)
	return i, ok
}

// GetSessionTimeSync returns the session time sync.
func GetSessionTimeSync(ctx context.Context) (enabled bool, ok bool) {
	d, ok := getState(ctx, func(st *state) interface{} {
		if st.TimeSync != nil {
			return *st.TimeSync
		}
		return nil
	}).(bool)
	return d, ok
}
