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

package test

import (
	"math/rand"

	"go.thethings.network/lorawan-stack/pkg/util/randutil"
)

// Randy is global rand, which is (mostly) safe for concurrent use.
// Read and Seed should not be called concurrently.
var Randy = rand.New(randutil.NewLockedSource(rand.NewSource(42)))

func init() {
	rand.Seed(42)
}
