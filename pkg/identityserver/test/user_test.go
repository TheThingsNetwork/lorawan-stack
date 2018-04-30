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

package test

import (
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func user() *ttnpb.User {
	return &ttnpb.User{
		UserIdentifiers: ttnpb.UserIdentifiers{
			UserID: "alice",
			Email:  "alice@alice.com",
		},
		Name:              "Ali Ce",
		Password:          "123456",
		ValidatedAt:       timeValue(now),
		PasswordUpdatedAt: now,
	}
}

func TestShouldBeUser(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeUser(user(), user()), should.Equal, success)

	modified := user()
	modified.CreatedAt = time.Now()

	a.So(ShouldBeUser(modified, user()), should.NotEqual, success)
}

func TestShouldBeUserIgnoringAutoFields(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeUserIgnoringAutoFields(user(), user()), should.Equal, success)

	modified := user()
	modified.Name = "foo"

	a.So(ShouldBeUserIgnoringAutoFields(modified, user()), should.NotEqual, success)
}
