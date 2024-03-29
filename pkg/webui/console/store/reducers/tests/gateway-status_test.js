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

import reducer, { defaultState } from '../gateway-status'
import {
  getGateway,
  updateGatewayStatisticsSuccess,
  getGatewayEventMessageSuccess,
} from '../../actions/gateways'

describe('Gateway-status reducer', () => {
  it('returns the initial state', () => {
    expect(reducer(undefined, { type: '@@TEST_INIT', payload: {} })).toEqual(defaultState)
  })

  describe('when receiving stats update', () => {
    let newState = null

    it('sets `lastSeen` from `last_status_received_at` if present', () => {
      const stats = {
        last_status_received_at: '2019-09-24T13:40:30.033728431Z',
        last_uplink_received_at: '2019-09-24T13:40:10.536603866Z',
      }

      newState = reducer(defaultState, updateGatewayStatisticsSuccess({ stats }))

      expect(newState !== defaultState).toBe(true)
      expect(newState.lastSeen).toBeDefined()
      expect(newState.lastSeen).toEqual(stats.last_status_received_at)
    })

    it('sets `lastSeen` from `last_status_received_at` if have more recent value', () => {
      const stats = {
        last_status_received_at: '2019-09-24T13:35:10.033728431Z',
        last_uplink_received_at: '2019-09-24T13:35:30.536603866Z',
      }

      newState = reducer(newState, updateGatewayStatisticsSuccess({ stats }))

      expect(newState !== defaultState).toBe(true)
      expect(newState.lastSeen).toBeDefined()
      expect(newState.lastSeen).not.toEqual(stats.last_uplink_received_at)
    })

    it("updates `lastSeen` from `last_uplink_received_at` if it's the most recent value", () => {
      const stats = {
        last_status_received_at: '2019-09-24T13:45:30.033728431Z',
        last_uplink_received_at: '2019-09-24T13:40:10.536603866Z',
      }

      newState = reducer(newState, updateGatewayStatisticsSuccess({ stats }))

      expect(newState !== defaultState).toBe(true)
      expect(newState.lastSeen).toBeDefined()
      expect(newState.lastSeen).toEqual(stats.last_status_received_at)
    })

    it('sets `lastSeen` from `last_uplink_received_at` if `last_status_received_at` not present', () => {
      const stats = {
        last_uplink_received_at: '2019-09-24T13:50:10.536603866Z',
      }

      newState = reducer(newState, updateGatewayStatisticsSuccess({ stats }))

      expect(newState !== defaultState).toBe(true)
      expect(newState.lastSeen).toBeDefined()
      expect(newState.lastSeen).toEqual(stats.last_uplink_received_at)
    })

    it('sets `lastSeen` from `last_uplink_received_at` with most recent value if `last_status_received_at` not present', () => {
      const stats = {
        last_uplink_received_at: '2019-09-24T13:55:10.536603866Z',
      }

      newState = reducer(newState, updateGatewayStatisticsSuccess({ stats }))

      expect(newState !== defaultState).toBe(true)
      expect(newState.lastSeen).toBeDefined()
      expect(newState.lastSeen).toEqual(stats.last_uplink_received_at)
    })

    it('does not set `lastSeen` from `last_uplink_received_at` if have more recent value', () => {
      const stats = {
        last_uplink_received_at: '2019-09-24T13:35:10.536603866Z',
      }

      newState = reducer(newState, updateGatewayStatisticsSuccess({ stats }))

      expect(newState !== defaultState).toBe(true)
      expect(newState.lastSeen).toBeDefined()
      expect(newState.lastSeen).not.toEqual(stats.last_uplink_received_at)
    })

    it('resets state on gateway request', () => {
      newState = reducer(newState, getGateway('test-gtw-id'))

      expect(newState).toStrictEqual(defaultState)
    })
  })

  describe('receives gateway status events', () => {
    let newState = null

    it('updates `lastSeen` on uplink event', () => {
      const event = {
        name: 'gs.up.receive',
        time: '2019-09-24T13:40:30.033728431Z',
      }

      newState = reducer(defaultState, getGatewayEventMessageSuccess('test-gtw-id', event))
      expect(newState !== defaultState).toBe(true)
      expect(newState.lastSeen).toBeDefined()
      expect(newState.lastSeen).toStrictEqual(event.time)
    })

    it('updates `lastSeen` on most recent uplink event', () => {
      const event = {
        name: 'gs.up.receive',
        time: '2019-09-24T13:45:30.033728431Z',
      }

      newState = reducer(newState, getGatewayEventMessageSuccess('test-gtw-id', event))
      expect(newState !== defaultState).toBe(true)
      expect(newState.lastSeen).toBeDefined()
      expect(newState.lastSeen).toStrictEqual(event.time)
    })

    it('does not set `lastSeen` on uplink event if have more recent value', () => {
      const event = {
        name: 'gs.up.receive',
        time: '2019-09-24T13:30:30.033728431Z',
      }

      newState = reducer(newState, getGatewayEventMessageSuccess('test-gtw-id', event))
      expect(newState !== defaultState).toBe(true)
      expect(newState.lastSeen).toBeDefined()
      expect(newState.lastSeen).not.toStrictEqual(event.time)
    })

    it('updates `lastSeen` on status event', () => {
      const event = {
        name: 'gs.status.receive',
        time: '2019-09-24T13:50:30.033728431Z',
      }

      newState = reducer(newState, getGatewayEventMessageSuccess('test-gtw-id', event))
      expect(newState !== defaultState).toBe(true)
      expect(newState.lastSeen).toBeDefined()
      expect(newState.lastSeen).toStrictEqual(event.time)
    })

    it('updates `lastSeen` on most recent status event', () => {
      const event = {
        name: 'gs.status.receive',
        time: '2019-09-24T13:55:30.033728431Z',
      }

      newState = reducer(newState, getGatewayEventMessageSuccess('test-gtw-id', event))
      expect(newState !== defaultState).toBe(true)
      expect(newState.lastSeen).toBeDefined()
      expect(newState.lastSeen).toStrictEqual(event.time)
    })

    it('does not set `lastSeen` on status event if have more recent value', () => {
      const event = {
        name: 'gs.up.receive',
        time: '2019-09-24T13:35:30.033728431Z',
      }

      newState = reducer(newState, getGatewayEventMessageSuccess('test-gtw-id', event))
      expect(newState !== defaultState).toBe(true)
      expect(newState.lastSeen).toBeDefined()
      expect(newState.lastSeen).not.toStrictEqual(event.time)
    })

    it('resets state on gateway request', () => {
      newState = reducer(newState, getGateway('test-gtw-id'))

      expect(newState).toStrictEqual(defaultState)
    })
  })
})
