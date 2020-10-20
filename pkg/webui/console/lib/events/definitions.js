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

import { defineMessages } from 'react-intl'

import { defineSyntheticEvent } from './utils'

export const eventMessages = defineMessages({
  'synthetic.error.unknown:type': 'Unknown error',
  'synthetic.error.unknown:preview':
    'An unknown error occurred and one or more events could not be retrieved',

  'synthetic.error.network_error:type': 'Network error',
  'synthetic.error.network_error:preview': 'The stream connection was lost due to a network error',

  'synthetic.status.reconnecting:type': 'Reconnecting',
  'synthetic.status.reconnecting:preview': 'Attempting to reconnect…',

  'synthetic.status.reconnected:type': 'Stream reconnected',
  'synthetic.status.reconnected:preview': 'The stream connection has been re-established',

  'synthetic.status.closed:type': 'Stream connection closed',
  'synthetic.status.closed:preview': 'The connection was closed by the stream provider',

  'synthetic.status.cleared:type': 'Events cleared',
  'synthetic.status.cleared:preview': 'The events list has been cleared',

  'synthetic.status.paused:type': 'Stream paused',
  'synthetic.status.paused:preview': 'The event stream has been paused',

  'synthetic.status.resumed:type': 'Stream resumed',
  'synthetic.status.resumed:preview': 'The event stream has been resumed after being paused',
})

export const EVENT_UNKNOWN_ERROR = 'synthetic.error.unknown'
export const EVENT_NETWORK_ERROR = 'synthetic.error.network_error'
export const EVENT_STATUS_RECONNECTING = 'synthetic.status.reconnecting'
export const EVENT_STATUS_RECONNECTED = 'synthetic.status.reconnected'
export const EVENT_STATUS_CLOSED = 'synthetic.status.closed'
export const EVENT_STATUS_CLEARED = 'synthetic.status.cleared'
export const EVENT_STATUS_PAUSED = 'synthetic.status.paused'
export const EVENT_STATUS_RESUMED = 'synthetic.status.resumed'

export const createUnknownErrorEvent = defineSyntheticEvent(EVENT_UNKNOWN_ERROR)
export const createNetworkErrorEvent = defineSyntheticEvent(EVENT_NETWORK_ERROR)
export const createStatusReconnectingEvent = defineSyntheticEvent(EVENT_STATUS_RECONNECTING)
export const createStatusReconnectedEvent = defineSyntheticEvent(EVENT_STATUS_RECONNECTED)
export const createStatusClosedEvent = defineSyntheticEvent(EVENT_STATUS_CLOSED)
export const createStatusClearedEvent = defineSyntheticEvent(EVENT_STATUS_CLEARED)
export const createStatusPausedEvent = defineSyntheticEvent(EVENT_STATUS_PAUSED)
export const createStatusResumedEvent = defineSyntheticEvent(EVENT_STATUS_RESUMED)
