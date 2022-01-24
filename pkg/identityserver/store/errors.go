// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import "go.thethings.network/lorawan-stack/v3/pkg/errors"

var (
	// ErrIDTaken is returned when an entity can not be created because the ID is already taken.
	ErrIDTaken = errors.DefineAlreadyExists("id_taken", "ID already taken, choose a different one and try again")
	// ErrEUITaken is returned when an entity can not be created because the EUI is already taken.
	ErrEUITaken = errors.DefineAlreadyExists("eui_taken", "EUI already taken")
)
