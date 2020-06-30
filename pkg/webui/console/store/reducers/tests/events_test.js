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

import createReducer from '../events'
import { getEventMessageSuccess } from '../../actions/events'

describe('Events reducer', () => {
  const testId = 'test'
  const reducer = createReducer(testId)
  const successActionCreator = getEventMessageSuccess(testId)

  describe('when adding new events', () => {
    it('keeps events sorted by `time`', () => {
      const testId = 'test'
      const testTime = '2019-03-28T13:18:13.376022Z'
      const testDate = new Date(testTime)

      const initialState = reducer(undefined, {})

      const event1 = { time: testTime }
      let newState = reducer(initialState, successActionCreator(testId, event1))

      expect(newState[testId].events).toHaveLength(1)
      expect(newState[testId].events[0]).toEqual(event1)

      const event2 = { time: new Date(testDate.getTime() + 1000).toISOString() }
      newState = reducer(newState, successActionCreator(testId, event2))

      expect(newState[testId].events).toHaveLength(2)
      expect(newState[testId].events[0]).toEqual(event2)
      expect(newState[testId].events[1]).toEqual(event1)

      const event3 = { time: new Date(testDate.getTime() - 1000).toISOString() }
      newState = reducer(newState, successActionCreator(testId, event3))

      expect(newState[testId].events).toHaveLength(3)
      expect(newState[testId].events[0]).toEqual(event2)
      expect(newState[testId].events[1]).toEqual(event1)
      expect(newState[testId].events[2]).toEqual(event3)

      const event4 = { time: new Date(testDate.getTime() + 2000).toISOString() }
      newState = reducer(newState, successActionCreator(testId, event4))

      expect(newState[testId].events).toHaveLength(4)
      expect(newState[testId].events[0]).toEqual(event4)
      expect(newState[testId].events[1]).toEqual(event2)
      expect(newState[testId].events[2]).toEqual(event1)
      expect(newState[testId].events[3]).toEqual(event3)

      const event5 = { time: new Date(testDate.getTime()).toISOString() }
      newState = reducer(newState, successActionCreator(testId, event5))

      expect(newState[testId].events).toHaveLength(5)
      expect(newState[testId].events[0]).toEqual(event4)
      expect(newState[testId].events[1]).toEqual(event2)
      expect(newState[testId].events[2]).toEqual(event5)
      expect(newState[testId].events[3]).toEqual(event1)
      expect(newState[testId].events[4]).toEqual(event3)
    })
  })
})
