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

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { defineSyntheticEvent } from './utils'

export const eventMessages = {
  'synthetic.error.unknown:type': sharedMessages.unknownError,
  ...defineMessages({
    'synthetic.error.unknown:preview':
      'An unknown error occurred and one or more events could not be retrieved',

    'synthetic.error.unknown:details':
      'The Console encountered an unexpected error while handling the event stream data. It is possible that event data could not be displayed (correctly) as a result. Note that this is an internal error which does not imply any malfunction of your gateways or end devices.',

    'synthetic.error.network_error:type': 'Network error',

    'synthetic.error.network_error:preview':
      'The stream connection was lost due to a network error',

    'synthetic.error.network_error:details':
      'The Console was not able to fetch further stream events because the network connection of your host machine was interrupted. This can have various causes, such as your host machine switching Wi-Fi networks or experiencing drops in signal strength. Please check your internet connection and ensure a stable internet connection to avoid stream disconnects. The stream will reconnect automatically once the internet connection has been re-established.',

    'synthetic.status.reconnecting:type': 'Reconnecting',
    'synthetic.status.reconnecting:preview': 'Attempting to reconnect…',

    'synthetic.status.reconnecting:details':
      'The Console will periodically try to reconnect to the event stream if the connection was interrupted.',

    'synthetic.status.reconnected:type': 'Stream reconnected',
    'synthetic.status.reconnected:preview': 'The stream connection has been re-established',
    'synthetic.status.reconnected:details':
      'The Console was able to reconnect to the network and resumed the event stream. Subsequent event data will be received and displayed. Note that event data which was possibly emitted during the network disruption will only be re-delivered within the range of the event retention configuration for your network.',

    'synthetic.status.closed:type': 'Stream connection closed',
    'synthetic.status.closed:preview': 'The connection was closed',
    'synthetic.status.closed:details':
      'The event stream was closed. This can be for various reasons, such as your machine going into standby mode or the internet connection being interrupted. The stream will reconnect automatically once the internet connection has been re-established.',

    'synthetic.status.cleared:type': 'Events cleared',
    'synthetic.status.cleared:preview': 'The events list has been cleared',
    'synthetic.status.cleared:details': 'The list of displayed events has been cleared.',
    'synthetic.status.paused:type': 'Stream paused',
    'synthetic.status.paused:preview': 'The event stream has been paused',

    'synthetic.status.paused:details':
      'The event stream has been paused by the user. Subsequent event data will not be displayed until the stream is resumed.',

    'synthetic.status.resumed:type': 'Stream resumed',
    'synthetic.status.resumed:preview': 'The event stream has been resumed after being paused',

    'synthetic.status.resumed:details':
      'The event stream has been resumed by the user and will receive new subsequent event data. Note that event data which was possibly emitted during the stream pause will not be re-delivered.',

    'synthetic.status.filter_enabled:type': 'Filter enabled',
    'synthetic.status.filter_enabled:preview': 'An event filter has been enabled',

    'synthetic.status.filter_enabled:details':
      'The event stream will now be filtered, meaning that some event types may be suppressed. See the details below for more information about the currently enabled filter.',

    'synthetic.status.filter_disabled:type': 'Filter disabled',
    'synthetic.status.filter_disabled:preview': 'The previously set event filter has been disabled',

    'synthetic.status.filter_disabled:details':
      'The event stream will not be filtered anymore. This means that you will see all events that come out of the event stream for this entity.',
  }),
}

export const EVENT_UNKNOWN_ERROR = 'synthetic.error.unknown'
export const EVENT_NETWORK_ERROR = 'synthetic.error.network_error'
export const EVENT_STATUS_RECONNECTING = 'synthetic.status.reconnecting'
export const EVENT_STATUS_RECONNECTED = 'synthetic.status.reconnected'
export const EVENT_STATUS_CLOSED = 'synthetic.status.closed'
export const EVENT_STATUS_CLEARED = 'synthetic.status.cleared'
export const EVENT_STATUS_PAUSED = 'synthetic.status.paused'
export const EVENT_STATUS_RESUMED = 'synthetic.status.resumed'
export const EVENT_STATUS_FILTER_ENABLED = 'synthetic.status.filter_enabled'
export const EVENT_STATUS_FILTER_DISABLED = 'synthetic.status.filter_disabled'

export const createUnknownErrorEvent = defineSyntheticEvent(EVENT_UNKNOWN_ERROR)
export const createNetworkErrorEvent = defineSyntheticEvent(EVENT_NETWORK_ERROR)
export const createStatusReconnectingEvent = defineSyntheticEvent(EVENT_STATUS_RECONNECTING)
export const createStatusReconnectedEvent = defineSyntheticEvent(EVENT_STATUS_RECONNECTED)
export const createStatusClosedEvent = defineSyntheticEvent(EVENT_STATUS_CLOSED)
export const createStatusClearedEvent = defineSyntheticEvent(EVENT_STATUS_CLEARED)
export const createStatusPausedEvent = defineSyntheticEvent(EVENT_STATUS_PAUSED)
export const createStatusResumedEvent = defineSyntheticEvent(EVENT_STATUS_RESUMED)
export const createStatusFilterEnabled = defineSyntheticEvent(EVENT_STATUS_FILTER_ENABLED)
export const createStatusFilterDisabled = defineSyntheticEvent(EVENT_STATUS_FILTER_DISABLED)
