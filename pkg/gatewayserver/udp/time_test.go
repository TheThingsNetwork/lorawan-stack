// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
