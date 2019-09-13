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

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestPopulate(t *testing.T) {
	a := assertions.New(t)
	p := NewPopulator(10, 42)

	a.So(p.Users, should.HaveLength, 10)

	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, models...)
		err := p.Populate(ctx, db)

		a.So(err, should.BeNil)
	})
}
