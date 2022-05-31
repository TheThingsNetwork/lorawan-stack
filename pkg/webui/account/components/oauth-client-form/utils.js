// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

export const approvalStates = [
  'STATE_REQUESTED',
  'STATE_APPROVED',
  'STATE_REJECTED',
  'STATE_FLAGGED',
  'STATE_SUSPENDED',
]

export const encodeGrants = value => {
  const grants = Object.keys(value).map(grant => {
    if (value[grant]) {
      return grant
    }

    return null
  })

  return grants.filter(Boolean)
}

export const decodeGrants = value => {
  if (value) {
    const grants = value.reduce((g, i) => {
      g[i] = true
      return g
    }, {})

    return grants
  }

  return null
}
