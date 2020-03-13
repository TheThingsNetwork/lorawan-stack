// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import Event from '..'

const getEventActionByName = function(name) {
  const names = name.split('.')

  return names[names.length - 1]
}

const getEventDataType = function(type) {
  const entries = type.split('.')

  return entries[entries.length - 1]
}

const isErrorEvent = function(data) {
  if (!data) {
    return false
  }

  const { '@type': t } = data
  return getEventDataType(t) === 'ErrorDetails'
}

const getEventComponentByName = function(event) {
  const { name, data } = event
  const action = getEventActionByName(name)

  let component = null
  let type = null
  if (isErrorEvent(data)) {
    component = Event.Error
  } else if (name.includes('.up.')) {
    component = Event.Message
    type = 'uplink'
  } else if (name.includes('.down.')) {
    component = Event.Message
    type = 'downlink'
  } else if (['create', 'delete', 'update'].includes(action)) {
    component = Event.CRUD
    type = 'crud'
  } else {
    component = Event.Default
    type = 'default'
  }

  return { component, type }
}

const formatMessageData = function(data) {
  if (!data) {
    return null
  }

  const { '@type': t, ...rest } = data
  const type = getEventDataType(t)

  switch (type) {
    case 'UplinkMessage':
    case 'DownlinkMessage':
    case 'ApplicationUp':
    case 'ErrorDetails':
      return rest
    default:
      return null
  }
}

export {
  getEventComponentByName as default,
  getEventActionByName,
  getEventDataType,
  formatMessageData,
}
