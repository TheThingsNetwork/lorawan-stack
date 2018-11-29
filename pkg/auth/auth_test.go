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

package auth_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestAuth(t *testing.T) {
	a := assertions.New(t)

	token, err := auth.APIKey.Generate(test.Context(), "")
	a.So(err, should.BeNil)

	tokenType, id, key, err := auth.SplitToken(token)
	a.So(err, should.BeNil)
	a.So(tokenType, should.Equal, auth.APIKey)

	a.So(auth.JoinToken(tokenType, id, key), should.Equal, token)

	for _, token := range []string{
		"FOO",             // invalid length
		"FOO.FOO",         // invalid length
		"FOO.FOO.FOO",     // invalid type
		"FOO.FOO.FOO.FOO", // invalid length
	} {
		_, _, _, err := auth.SplitToken(token)
		a.So(err, should.NotBeNil)
	}
}
