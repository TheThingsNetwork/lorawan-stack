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

const getLineNumberOfIndex = (data, index) => {
  const lines = data.split('\n')
  let line = 0
  let i = 0

  while (i < index) {
    i += lines[line].length + 1
    line++
  }

  return line
}

const insertAtLineAtIndex = (oldContents, index, content) => {
  const line = getLineNumberOfIndex(oldContents, index)
  const lines = oldContents.split('\n')
  lines.splice(line, 0, content)

  return `${lines
    .join('\n')
    .replace(/\n{3,}/g, '\n\n')
    .replace(/\n*$/, '')}`
}

export { insertAtLineAtIndex }
