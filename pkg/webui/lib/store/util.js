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

/* eslint-disable import/prefer-default-export */

import { cloneDeepWith } from 'lodash'

export const trimEvents = state => ({
  ...state,
  events: cloneDeepWith(state.events, (value, key) => {
    if (key === 'events' && value instanceof Array) {
      // Only transfer the last 5 events to Sentry to avoid
      // `Payload too large` errors.
      return value.slice(0, 5)
    }
  }),
})
