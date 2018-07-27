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

// Package cluster contains cluster authentication-related utilities.
package cluster

import "context"

type clusterAuthKeys string

var (
	clusterAuthKey        = clusterAuthKeys("key")
	clusterAuthFailureKey = clusterAuthKeys("failure")
)

// NewContext returns a context containing whether the
func NewContext(ctx context.Context, err error) context.Context {
	ctx = context.WithValue(ctx, clusterAuthKey, err == nil)
	ctx = context.WithValue(ctx, clusterAuthFailureKey, err)
	return ctx
}

// Authorized returns whether the context has been identified as a cluster call.
// It panics if it was not created with `NewContext`.
func Authorized(ctx context.Context) error {
	ok, isStored := ctx.Value(clusterAuthKey).(bool)
	if !isStored {
		panic("Tried to verify the call source of the context, but no call source was registered in the first place")
	}
	if ok {
		return nil
	}
	err := ctx.Value(clusterAuthFailureKey).(error)
	return err
}
