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

import { createNetworkErrorEvent, createUnknownErrorEvent } from './definitions'

export const defineSyntheticEvent = name => data => ({
  time: new Date().toISOString(),
  name,
  isError: name.startsWith('synthetic.error'),
  isSynthetic: true,
  unique_id: `synthetic.${Date.now()}`,
  data,
})

export const createSyntheticEventFromError = error => {
  if (error instanceof Error) {
    const errorString = error.toString()
    if (error.message === 'network error') {
      return createNetworkErrorEvent({ error: errorString })
    }

    return createUnknownErrorEvent({ error: errorString })
  }
}
