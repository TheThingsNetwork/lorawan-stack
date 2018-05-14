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

package db

import (
	"unicode"

	"github.com/jmoiron/sqlx"
)

func init() {
	// Set a custom NameMapper function in sqlx before it is used.
	// This custom NameMapper will be used by sqlx to construct mappers which
	// will match struct fields with database columns.
	sqlx.NameMapper = nameMapper
}

// nameMapper is the custom method used to map struct fields names to columns.
// For a struct field name it returns it's lowered case version by also placing
// underscore before non preceding uppercase characters.
//
// Examples:
//    RedirectURI -> redirect_uri
//    ArchivedAt  -> archived_at
//    UserID      -> user_id
func nameMapper(fieldName string) string {
	// The output string will at least have the same length as the input string.
	res := make([]byte, 0, len(fieldName))

	for i, char := range fieldName {
		if i != 0 && unicode.IsUpper(char) && !unicode.IsUpper(rune(fieldName[i-1])) && fieldName[i-1] != byte('_') {
			res = append(res, byte('_'))
		}
		res = append(res, byte(unicode.ToLower(char)))
	}

	return string(res)
}
