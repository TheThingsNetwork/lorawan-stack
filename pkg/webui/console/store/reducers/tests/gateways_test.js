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

import reducer from '../gateways'
import {
  getGateway,
  getGatewaySuccess,
  getGatewayFailure,
  getGatewaysList,
  getGatewaysListSuccess,
  getGatewaysListFailure,
  updateGateway,
  updateGatewaySuccess,
  updateGatewayFailure,
  deleteGateway,
  deleteGatewaySuccess,
  deleteGatewayFailure,
} from '../../actions/gateways'

describe('Gateways reducer', function () {
  const defaultState = {
    entities: {},
    selectedGateway: null,
    statistics: {},
  }

  it('returns the initial state', function () {
    expect(reducer(undefined, { type: '@@TEST_INIT', payload: {} })).toEqual(defaultState)
  })

  it('ignores `getGatewayFailure` action', function () {
    expect(reducer(defaultState, getGatewayFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('ignores `updateGatewayFailure` action', function () {
    expect(reducer(defaultState, updateGatewayFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('ignores `getGatewaysListFailure` action', function () {
    expect(reducer(defaultState, getGatewaysListFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('ignores `updateGateway` action', function () {
    expect(reducer(defaultState, updateGateway('test-id', {}))).toEqual(defaultState)
  })

  it('ignores `deleteGatewayFailure` action', function () {
    expect(reducer(defaultState, deleteGatewayFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('ignores `deleteGateway` action', function () {
    expect(reducer(defaultState, deleteGateway('test-id'))).toEqual(defaultState)
  })

  describe('when requesting a single gateway', function () {
    const testGatewayId = 'tesrt-gtw-id'
    const testGateway = { ids: { gateway_id: testGatewayId }, name: 'test-gtw-name' }
    let newState

    beforeAll(function () {
      newState = reducer(defaultState, getGateway(testGatewayId))
    })

    it('sets `selectedGateway` on `getGateway` action', function () {
      expect(newState.selectedGateway).toEqual(testGatewayId)
    })

    it('updates `entities` on `getGateway` action', function () {
      expect(newState.entities).toEqual(defaultState.entities)
    })

    describe('when it receives the gateway', function () {
      beforeAll(function () {
        newState = reducer(newState, getGatewaySuccess(testGateway))
      })

      it('does not change `selectedGateway` on `getGatewaySuccess` action', function () {
        expect(newState.selectedGateway).toEqual(testGatewayId)
      })

      it('adds new gateway to `entities` on `getGatewaySuccess` action', function () {
        expect(Object.keys(newState.entities)).toHaveLength(1)
        expect(newState.entities[testGatewayId]).toEqual(testGateway)
      })

      describe('when it updates the gateway', function () {
        const updatedTestGateway = { ids: { gateway_id: testGatewayId }, name: 'updated-test-gtw' }
        let updatedState

        beforeAll(function () {
          updatedState = reducer(newState, updateGatewaySuccess(updatedTestGateway))
        })

        it('does not change `selectedGateway` on `updateGatewaySuccess` action', function () {
          expect(updatedState.selectedGateway).toEqual(testGatewayId)
        })

        it('updates the gateway in `entities` on `updateGatewaySuccess` action', function () {
          expect(updatedState.entities[testGatewayId].name).toEqual(updatedTestGateway.name)
        })
      })

      describe('when deleting the gateway', function () {
        let updatedState

        beforeAll(function () {
          updatedState = reducer(newState, deleteGatewaySuccess({ id: testGatewayId }))
        })

        it('removes `selectedGateway` on `deleteGatewaySuccess` action', function () {
          expect(updatedState.selectedGateway).toBeNull()
        })

        it('removes gateway in `entities` on `deleteGatewaySuccess` action', function () {
          expect(updatedState.entities[testGatewayId]).toBeUndefined()
        })
      })

      describe('when requesting another gateway', function () {
        const otherTestGatewayId = 'another-test-gtw-id'
        const otherTestGateway = {
          ids: { gateway_id: otherTestGatewayId },
          name: 'another-test-gtw',
        }
        let updatedState

        beforeAll(function () {
          updatedState = reducer(newState, getGateway(otherTestGatewayId))
        })

        it('sets `selectedGateway` on `getGateway` action', function () {
          expect(updatedState.selectedGateway).toEqual(otherTestGatewayId)
        })

        it('does not update `entities` on `getGateway` action', function () {
          expect(Object.keys(updatedState.entities)).toHaveLength(1)
          expect(updatedState.entities[testGatewayId]).toEqual(testGateway)
        })

        describe('when receiving the gateway', function () {
          beforeAll(function () {
            updatedState = reducer(updatedState, getGatewaySuccess(otherTestGateway))
          })

          it('does not change `selectedGateway` on `getGatewaySuccess` action', function () {
            expect(updatedState.selectedGateway).toEqual(otherTestGatewayId)
          })

          it('keeps previously received gateway in `entities`', function () {
            expect(updatedState.entities[testGatewayId]).toEqual(testGateway)
          })

          it('adds new gateway to `entities` on `getGatewaySuccess`', function () {
            expect(Object.keys(updatedState.entities)).toHaveLength(2)
            expect(updatedState.entities[otherTestGatewayId]).toEqual(otherTestGateway)
          })
        })
      })

      describe('requesting a list of gateways', function () {
        beforeAll(function () {
          newState = reducer(newState, getGatewaysList({}))
        })

        it('does not change `selectedGateway` on `getGatewaysList` action', function () {
          expect(newState.selectedGateway).toEqual(testGatewayId)
        })

        it('does not change `entities` on `getGatewaysList` action', function () {
          expect(Object.keys(newState.entities)).toHaveLength(1)
          expect(newState.entities[testGatewayId]).toEqual(testGateway)
        })

        describe('receiving the list of gateways', function () {
          const entities = [
            { ids: { gateway_id: 'test-gtw-1' }, name: 'test-gtw-1' },
            { ids: { gateway_id: 'test-gtw-2' }, name: 'test-gtw-2' },
            { ids: { gateway_id: 'test-gtw-3' }, name: 'test-gtw-3' },
          ]
          const totalCount = entities.length

          beforeAll(function () {
            newState = reducer(newState, getGatewaysListSuccess({ entities, totalCount }))
          })

          it('does not remove previously received gateway on `getGatewaysListSuccess` action', function () {
            expect(newState.entities[testGatewayId]).toEqual(testGateway)
          })

          it('adds new gateways to `entities` on `getGatewaysListSuccess`', function () {
            expect(Object.keys(newState.entities)).toHaveLength(4)
            for (const gtw of entities) {
              expect(newState.entities[gtw.ids.gateway_id]).toEqual(gtw)
            }
          })
        })
      })
    })
  })
})
