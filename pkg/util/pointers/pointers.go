// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package pointers

import "time"

// Bool returns a boolean pointer for the given value.
func Bool(v bool) *bool { return &v }

// Duration returns a duration pointer for the given value.
func Duration(d time.Duration) *time.Duration { return &d }

// Time returns a time.Time pointer for the given value.
func Time(t time.Time) *time.Time { return &t }

// Float32 returns a float32 pointer for the given value.
func Float32(v float32) *float32 { return &v }

// Uint8 returns a uint8 pointer for the given value.
func Uint8(v uint8) *uint8 { return &v }
