// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package ttnpb

import (
	"time"

	pbtypes "github.com/gogo/protobuf/types"
)

// StdTime converts a protobuf timestamp to a standard library time.
//
// ProtoTime panics if the time is invalid.
func StdTime(protoTime *pbtypes.Timestamp) *time.Time {
	if protoTime == nil {
		return nil
	}
	stdTime, err := pbtypes.TimestampFromProto(protoTime)
	if err != nil {
		panic(err)
	}
	return &stdTime
}

// StdTimeOrZero converts a protobuf time to a standard library time.
// If protoTime is nil, it returns a zero time.
func StdTimeOrZero(protoTime *pbtypes.Timestamp) time.Time {
	stdTime := StdTime(protoTime)
	if stdTime == nil {
		return time.Time{}
	}
	return *stdTime
}

// ProtoTime converts a standard library time to a protobuf timestamp.
//
// ProtoTime panics if the time is invalid.
func ProtoTime(stdTime *time.Time) *pbtypes.Timestamp {
	if stdTime == nil {
		return nil
	}
	protoTime, err := pbtypes.TimestampProto(*stdTime)
	if err != nil {
		panic(err)
	}
	return protoTime
}

// ProtoTimePtr converts a standard library time to a pointer and then to a protobuf timestamp.
func ProtoTimePtr(stdTime time.Time) *pbtypes.Timestamp { return ProtoTime(&stdTime) }
