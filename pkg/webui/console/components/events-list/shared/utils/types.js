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

import * as eventsRegexp from './regexp'

// Generic utilities.

/**
 * Returns the name of the events payload.
 *
 * @param {object} event - The event object.
 * @returns {string} - The name of the events payload.
 */
const getEventDataType = event => {
  if (event && 'data' in event) {
    const type = event.data['@type']
    const entries = type.split('.')

    return entries[entries.length - 1]
  }
}

// Event types.

/**
 * Checks if `event` represents create, delete or update event of an entity.
 *
 * @param {object} event - The event object.
 * @returns {boolean} - `true` if `event` is the CRUD event, `false` otherwise.
 */
export const isCRUDEvent = event => {
  return eventsRegexp.crud.test(event.name)
}

/**
 * Checks if `event` represents the create event of an entity.
 *
 * @param {object} event - The event object.
 * @returns {boolean} - `true` if `event` is the create event, `false` otherwise.
 */
export const isCRUDCreateEvent = event => {
  return eventsRegexp.crudCreate.test(event.name)
}

/**
 * Checks if `event` represents the update event of an entity.
 *
 * @param {object} event - The event object.
 * @returns {boolean} - `true` if `event` is the update event, `false` otherwise.
 */
export const isCRUDUpdateEvent = event => {
  return eventsRegexp.crudUpdate.test(event.name)
}

/**
 * Checks if `event` represents the delete event of an entity.
 *
 * @param {object} event - The event object.
 * @returns {boolean} - `true` if `event` is the delete event, `false` otherwise.
 */
export const isCRUDDeleteEvent = event => {
  return eventsRegexp.crudDelete.test(event.name)
}

/**
 * Checks if `event` represents the end device uplink event.
 *
 * @param {object} event - The event object.
 * @returns {boolean} - `true` if `event` is the end device uplink event, `false` otherwise.
 */
export const isDeviceUplinkEvent = event => {
  return eventsRegexp.deviceUplink.test(event.name)
}

/**
 * Checks if `event` represents the end device downlink event.
 *
 * @param {object} event - The event object.
 * @returns {boolean} - `true` if `event` is the end device downlink event, `false` otherwise.
 */
export const isDeviceDownlinkEvent = event => {
  return eventsRegexp.deviceDownlink.test(event.name)
}

/**
 * Checks if `event` represents the end device join event.
 *
 * @param {object} event - The event object.
 * @returns {boolean} - `true` if `event` is the end device join event, `false` otherwise.
 */
export const isDeviceJoinEvent = event => {
  return eventsRegexp.deviceJoin.test(event.name)
}

export const isDeviceNsUtilityEvent = event => {
  return (
    !isDeviceUplinkEvent(event) &&
    !isDeviceDownlinkEvent(event) &&
    !isDeviceJoinEvent(event) &&
    eventsRegexp.ns.test(event.name)
  )
}

/**
 * Checks if `event` represents the gateway uplink event.
 *
 * @param {object} event - The event object.
 * @returns {boolean} - `true` if `event` is the gateway uplink event, `false` otherwise.
 */
export const isGatewayUplinkEvent = event => {
  return eventsRegexp.gatewayUplink.test(event.name)
}

/**
 * Checks if `event` represents the gateway downlink event.
 *
 * @param {object} event - The event object.
 * @returns {boolean} - `true` if `event` is the gateway downlink event, `false` otherwise.
 */
export const isGatewayDownlinkEvent = event => {
  return eventsRegexp.gatewayDownlink.test(event.name)
}

/**
 * Checks if `event` represents the error event.
 *
 * @param {object} event - The event object.
 * @returns {boolean} - `true` if `event` is the error event, `false` otherwise.
 */
export const isErrorEvent = event => {
  // Note: any event can be the error event regardless of its `name` field.
  return isErrorDetailsDataType(event)
}

export const isGatewayConnectionEvent = event => {
  return eventsRegexp.gatewayConnection.test(event.name)
}

export const isGatewayConnectEvent = event => {
  const match = event.name.match(eventsRegexp.gatewayConnection)

  return match[1] === 'connect'
}

export const isGatewayDisconnectEvent = event => {
  const match = event.name.match(eventsRegexp.gatewayConnection)

  return match[1] === 'disconnect'
}

// Event data types.

/**
 * Checks if `event` has an error as payload.
 *
 * @param {object} event - The event object.
 * @returns {boolean} - `true` if `event` has an error as payload, `false` otherwise.
 */
export const isErrorDetailsDataType = event => {
  const type = getEventDataType(event)

  return type === 'ErrorDetails'
}

export const isApplicationUpDataType = event => {
  const type = getEventDataType(event)

  return type === 'ApplicationUp'
}

export const isApplicationUplinkDataType = event => {
  const type = getEventDataType(event)

  return type === 'ApplicationUplink'
}

export const isJoinRequestDataType = event => {
  const type = getEventDataType(event)

  return type === 'JoinRequest'
}

export const isApplicationDownlinkDataType = event => {
  const type = getEventDataType(event)

  return type === 'ApplicationDownlink'
}

export const isUplinkMessageDataType = event => {
  const type = getEventDataType(event)

  return type === 'UplinkMessage'
}

export const isDownlinkMessageDataType = event => {
  const type = getEventDataType(event)

  return type === 'DownlinkMessage'
}

export const hasJoinAcceptData = event => {
  if (isApplicationUpDataType(event)) {
    const { data } = event

    return 'join_accept' in data
  }

  return false
}

export const hasUplinkMessageData = event => {
  if (isApplicationUpDataType(event)) {
    const { data } = event

    return 'uplink_message' in data
  }

  return false
}

export const hasJoinRequestData = event => {
  if (
    isApplicationUplinkDataType(event) ||
    isJoinRequestDataType(event) ||
    isUplinkMessageDataType(event)
  ) {
    const { data } = event

    return 'payload' in data && 'join_request_payload' in data.payload
  }

  return false
}

export const hasMacData = event => {
  if (isApplicationUplinkDataType(event) || isUplinkMessageDataType(event)) {
    const { data } = event

    return 'payload' in data && 'mac_payload' in data.payload
  }

  return false
}
