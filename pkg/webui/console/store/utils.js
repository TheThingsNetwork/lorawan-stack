// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

// Update a range of values in an array by using another array and a start index.
export const fillIntoArray = (array, start, values, totalCount) => {
  const newArray = [...array]
  const end = Math.min(start + values.length, totalCount)
  for (let i = start; i < end; i++) {
    newArray[i] = values[i - start]
  }

  return newArray
}

export const pageToIndices = (page, limit) => {
  const startIndex = (page - 1) * limit
  const stopIndex = page * limit - 1

  return [startIndex, stopIndex]
}
