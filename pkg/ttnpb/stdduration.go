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

	durationpb "google.golang.org/protobuf/types/known/durationpb"
)

// StdDuration converts a protobuf duration to a standard library duration.
//
// ProtoDuration panics if the Duration is invalid.
func StdDuration(protoDuration *durationpb.Duration) *time.Duration {
	if protoDuration == nil {
		return nil
	}
	stdDuration := protoDuration.AsDuration()
	return &stdDuration
}

// StdDurationOrZero converts a protobuf duration to a standard library duration.
// If protoDuration is nil, it returns a zero duration.
func StdDurationOrZero(protoDuration *durationpb.Duration) time.Duration {
	stdDuration := StdDuration(protoDuration)
	if stdDuration == nil {
		return 0
	}
	return *stdDuration
}

// ProtoDuration converts a standard library duration to a protobuf duration.
func ProtoDuration(stdDuration *time.Duration) *durationpb.Duration {
	if stdDuration == nil {
		return nil
	}
	return durationpb.New(*stdDuration)
}
