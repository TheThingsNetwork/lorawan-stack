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

// Package version contains version and build variables set by the CI process.
package version

import (
	"fmt"
)

// String returns the version string.
func String() string {
	version := TTN
	if GitCommit != "" && BuildDate != "" {
		version += fmt.Sprintf(" (%s, %s)", GitCommit, BuildDate)
	}
	return version
}
