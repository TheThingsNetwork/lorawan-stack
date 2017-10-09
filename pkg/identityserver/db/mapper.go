// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"unicode"

	"github.com/gomezjdaniel/sqlx"
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
//    CallbackURI -> callback_uri
//    ArchivedAt  -> archived_at
//    UserID      -> user_id
func nameMapper(fieldName string) string {
	// at least the output string will have the same length as the input string
	res := make([]byte, 0, len(fieldName))

	for i, char := range fieldName {
		if i != 0 && unicode.IsUpper(char) && !unicode.IsUpper(rune(fieldName[i-1])) && fieldName[i-1] != byte('_') {
			res = append(res, byte('_'))
		}
		res = append(res, byte(unicode.ToLower(char)))
	}

	return string(res)
}
