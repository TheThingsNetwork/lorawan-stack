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

import * as Integrations from '@sentry/integrations'

import env from '@ttn-lw/lib/env'

const sentryConfig = {
  dsn: env.sentryDsn,
  release: process.env.VERSION,
  normalizeDepth: 10,
  integrations: [new Integrations.Dedupe()],
  beforeSend: event => {
    if (event.extra.state && event.extra.state.user) {
      delete event.extra.state.user.user.name
    }
    return event
  },
}

export default sentryConfig
