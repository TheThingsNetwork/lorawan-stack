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

package redis

import (
	"github.com/go-redis/redis"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

var (
	errNotFound            = errors.DefineNotFound("not_found", "entity not found")
	errStore               = errors.Define("store", "store error")
	errInvalidKeyValueType = errors.DefineInvalidArgument("value_type", "invalid value type for key `{key}`")
	errMissingKey          = errors.DefineInvalidArgument("missing_key", "missing key `{key}`")
)

// ConvertError converts Redis error into errors.Error.
func ConvertError(err error) error {
	switch err {
	case nil:
		return nil
	case redis.Nil:
		return errNotFound
	default:
		return errStore.WithCause(err)
	}
}
