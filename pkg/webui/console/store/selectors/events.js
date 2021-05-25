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

import CONNECTION_STATUS from '@console/constants/connection-status'

const selectEventsStore = (state, entityId) => state[entityId] || {}

export const createEventsSelector = entity => (state, entityId) => {
  const store = selectEventsStore(state.events[entity], entityId)

  return store.events || []
}

export const createEventsStatusSelector = entity => (state, entityId) => {
  const store = selectEventsStore(state.events[entity], entityId)

  return store.status || CONNECTION_STATUS.UNKNOWN
}

export const createEventsPausedSelector = entity => (state, entityId) => {
  const store = selectEventsStore(state.events[entity], entityId)

  return Boolean(store.paused)
}

export const createEventsInterruptedSelector = entity => (state, entityId) => {
  const store = selectEventsStore(state.events[entity], entityId)

  return Boolean(store.interrupted)
}

export const createEventsErrorSelector = entity => (state, entityId) => {
  const store = selectEventsStore(state.events[entity], entityId)

  return store.error
}

export const createEventsTruncatedSelector = entity => (state, entityId) => {
  const store = selectEventsStore(state.events[entity], entityId)

  return Boolean(store.truncated)
}

export const createLatestEventSelector = entity => {
  const eventsSelector = createEventsSelector(entity)

  const selectLatestEvent = (state, entityId, includeSynthetic = false) => {
    const events = eventsSelector(state, entityId)

    return includeSynthetic ? events[0] : events.find(e => !e.isSynthetic)
  }

  return selectLatestEvent
}

export const createInterruptedStreamsSelector = entity => state => {
  const eventsStore = state.events[entity]

  return Object.keys(eventsStore).reduce((acc, id) => {
    if (eventsStore[id].interrupted) {
      acc[id] = eventsStore[id]
    }

    return acc
  }, {})
}

export const createEventsFilterSelector = entity => (state, entityId) => {
  const store = selectEventsStore(state.events[entity], entityId)

  return store.filter
}
