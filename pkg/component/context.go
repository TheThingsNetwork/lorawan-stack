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

package component

import (
	"context"
	"time"
)

type crossContext struct {
	cancelCtx context.Context
	valueCtx  context.Context
}

// Deadline implements context.Context using the cancel context.
func (ctx *crossContext) Deadline() (deadline time.Time, ok bool) {
	return ctx.cancelCtx.Deadline()
}

// Done implements context.Context using the cancel context.
func (ctx *crossContext) Done() <-chan struct{} {
	return ctx.cancelCtx.Done()
}

// Err implements context.Context using the cancel context.
func (ctx *crossContext) Err() error {
	return ctx.cancelCtx.Err()
}

// Value implements context.Context using the value context.
func (ctx *crossContext) Value(key any) any {
	return ctx.valueCtx.Value(key)
}
