// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package errors

// Resemble returns true iff the given errors resemble, meaning that the Code,
// Namespace and Name of the errors are equal. A nil error only resembles nil.
// Invalid errors or definitions (including typed nil) never resemble anything.
func Resemble(a, b error) bool {
	if a == nil && b == nil {
		return true
	}
	ttnA, ok := From(a)
	if !ok {
		return false
	}
	ttnB, ok := From(b)
	if !ok {
		return false
	}
	return ttnA.Code() == ttnB.Code() &&
		ttnA.Namespace() == ttnB.Namespace() &&
		ttnA.Name() == ttnB.Name()
}
