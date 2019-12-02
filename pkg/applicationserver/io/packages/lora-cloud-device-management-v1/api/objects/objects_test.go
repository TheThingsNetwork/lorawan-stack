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

package objects_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/packages/lora-cloud-device-management-v1/api/objects"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestRequestPolymorphism(t *testing.T) {
	for _, tc := range []struct {
		provided string
		expected objects.Request
	}{
		{
			provided: `{
				"type": "RESET",
				"param": 123
			}`,
			expected: objects.Request{
				Type:  objects.ResetRequestType,
				Param: ResetRequestParam(123),
			},
		},
		{
			provided: `{
				"type": "REJOIN"
			}`,
			expected: objects.Request{
				Type: objects.RejoinRequestType,
			},
		},
		{
			provided: `{
				"type": "MUTE"
			}`,
			expected: objects.Request{
				Type: objects.MuteRequestType,
			},
		},
		{
			provided: `{
				"type": "GETINFO",
				"param": ["temp", "adrmode", "interval"]
			}`,
			expected: objects.Request{
				Type: objects.GetInfoRequestType,
				Param: &objects.GetInfoRequestParam{
					"temp", "adrmode", "interval",
				},
			},
		},
		{
			provided: `{
				"type": "SETCONF",
				"param": {
					"adrmode": 12,
					"joineui": "80-00-00-00-00-00-00-0C",
					"interval": 23,
					"region": 34,
					"opmode": 456
				}
			}`,
			expected: objects.Request{
				Type: objects.SetConfRequestType,
				Param: &objects.SetConfRequestParam{
					ADRMode:  Uint8(12),
					JoinEUI:  objects.EUI{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C},
					Interval: Uint8(23),
					Region:   Uint8(34),
					OpMode:   Uint32(456),
				},
			},
		},
		{
			provided: `{
				"type": "FILEDONE",
				"param": {
					"sid": 123,
					"sctr": 234
				}
			}`,
			expected: objects.Request{
				Type: objects.FileDoneRequestType,
				Param: &objects.FileDoneRequestParam{
					SID:  123,
					SCtr: 234,
				},
			},
		},
		{
			provided: `{
				"type": "FUOTA",
				"param": "1223"
			}`,
			expected: objects.Request{
				Type:  objects.FUOTARequestType,
				Param: &objects.FUOTARequestParam{0x12, 0x23},
			},
		},
	} {
		t.Run(fmt.Sprintf("%v", tc.expected.Type), func(t *testing.T) {
			a := assertions.New(t)

			var request objects.Request
			err := json.Unmarshal([]byte(tc.provided), &request)
			a.So(err, should.BeNil)
			a.So(request, should.Resemble, tc.expected)
		})
	}
}
