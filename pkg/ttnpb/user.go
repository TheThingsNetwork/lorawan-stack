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

package ttnpb

import "regexp"

// GetUser returns the base User itself.
func (u *User) GetUser() *User {
	return u
}

var (
	// FieldPathUserName is the field path for the user name field.
	FieldPathUserName = regexp.MustCompile(`^name$`)

	// FieldPathUserEmail is the field path for the user email field.
	FieldPathUserEmail = regexp.MustCompile(`^ids.email$`)

	// FieldPathUserState is the field path for the user state field.
	FieldPathUserState = regexp.MustCompile(`^state$`)

	// FieldPathUserAdmin is the field path for the user admin field.
	FieldPathUserAdmin = regexp.MustCompile(`^admin$`)
)
