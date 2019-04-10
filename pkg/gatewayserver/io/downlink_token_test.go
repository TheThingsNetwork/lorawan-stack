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

package io

import (
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestDownlinkTokens(t *testing.T) {
	a := assertions.New(t)
	tokens := DownlinkTokens{}

	all := []uint16{}
	for i := 0; i < downlinkTokenItems*2; i++ {
		cids := []string{fmt.Sprintf("message_%d", i)}
		all = append(all, tokens.Next(cids, time.Unix(int64(i), 0)))

		for j, token := range all {
			cids, delta, ok := tokens.Get(token, time.Unix(int64(i), 0))
			if i-j < downlinkTokenItems {
				if !a.So(ok, should.BeTrue) {
					t.FailNow()
				}
				a.So(cids, should.Resemble, []string{fmt.Sprintf("message_%d", j)})
				a.So(delta, should.Equal, time.Duration(i-j)*time.Second)
			} else {
				a.So(ok, should.BeFalse)
			}
		}
	}
}
