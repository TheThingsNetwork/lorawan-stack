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

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestDownlinkTokens(t *testing.T) {
	a := assertions.New(t)
	tokens := DownlinkTokens{}

	msgs := make([]*ttnpb.DownlinkMessage, 0, downlinkTokenItems*2)
	all := []uint16{}
	for i := 0; i < downlinkTokenItems*2; i++ {

		msgs = append(msgs, &ttnpb.DownlinkMessage{
			RawPayload:     []byte{byte(i)},
			CorrelationIds: []string{fmt.Sprintf("corr_%d", i)},
		})
		all = append(all, tokens.Next(msgs[i], time.Unix(int64(i), 0)))

		for j, token := range all {
			msg, delta, ok := tokens.Get(token, time.Unix(int64(i), 0))
			if i-j < downlinkTokenItems {
				if !a.So(ok, should.BeTrue) {
					t.FailNow()
				}
				a.So(msg, should.Resemble, msgs[j])
				a.So(delta, should.Equal, time.Duration(i-j)*time.Second)
			} else {
				a.So(ok, should.BeFalse)
			}
		}
	}
}

func TestCorrelationIDs(t *testing.T) {
	tokens := DownlinkTokens{}

	msg := &ttnpb.DownlinkMessage{
		RawPayload:     []byte{0x00, 0x01},
		CorrelationIds: []string{"cid_downlink"},
	}
	token := tokens.Next(msg, time.Now())
	cid := tokens.FormatCorrelationID(token)
	parsedToken, ok := tokens.ParseTokenFromCorrelationIDs([]string{"cid:before", cid, "cid:after"})

	a := assertions.New(t)
	a.So(ok, should.BeTrue)
	a.So(parsedToken, should.Equal, token)

	matched, _, ok := tokens.Get(parsedToken, time.Now())
	a.So(ok, should.BeTrue)
	a.So(matched, should.Resemble, msg)

	_, ok = tokens.ParseTokenFromCorrelationIDs([]string{"cid1", "cid2"})
	a.So(ok, should.BeFalse)
}
