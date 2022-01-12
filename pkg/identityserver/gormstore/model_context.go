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

package store

import "context"

// SetContext needs to be called before creating models.
func (m *Model) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// SetContext sets the context on the organization model and the embedded account model.
func (org *Organization) SetContext(ctx context.Context) {
	org.Model.SetContext(ctx)
	org.Account.SetContext(ctx)
}

// SetContext sets the context on both the Model and Account.
func (usr *User) SetContext(ctx context.Context) {
	usr.Model.SetContext(ctx)
	usr.Account.SetContext(ctx)
}
