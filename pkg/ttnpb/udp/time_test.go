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

package udp

import (
	"testing"
	"time"
)

func TestCompactTime(t *testing.T) {
	nowTime := CompactTime(time.Now())
	_, err := nowTime.MarshalJSON()
	if err != nil {
		t.Error("Failed to marshal CompactTime to JSON:", err)
	}

	testTime := `"2017-06-06T15:04:05.999999Z"`
	testInvalidTime := "not a time string"

	var ct CompactTime
	err = ct.UnmarshalJSON([]byte(testInvalidTime))
	if err == nil {
		t.Error("An invalid CompactTime was accepted by the parser")
	}
	err = ct.UnmarshalJSON([]byte(testTime))
	if err != nil {
		t.Error("Failed to unmarshal CompactTime:", err)
	}
}

func TestExpandedTime(t *testing.T) {
	nowTime := ExpandedTime(time.Now())
	_, err := nowTime.MarshalJSON()
	if err != nil {
		t.Error("Failed to marshal ExpandedTime to JSON:", err)
	}

	testTime := `"2017-06-08 10:00:10 GMT"`
	testInvalidTime := "not a time string"

	var et ExpandedTime
	err = et.UnmarshalJSON([]byte(testInvalidTime))
	if err == nil {
		t.Error("An invalid ExpandedTime was accepted by the parser")
	}
	err = et.UnmarshalJSON([]byte(testTime))
	if err != nil {
		t.Error("Failed to unmarshal ExpandedTime:", err)
	}
}
