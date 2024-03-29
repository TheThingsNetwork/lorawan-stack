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

package ttnmage

import (
	"fmt"

	"github.com/magefile/mage/mg"
)

var initDeps []any

// Init initializes the tooling.
func Init() {
	mg.Deps(initDeps...)
}

func targetError(err error) error {
	if err != nil {
		return fmt.Errorf("failed checking modtime: %w", err)
	}
	return nil
}
