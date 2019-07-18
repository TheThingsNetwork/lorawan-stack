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

describe('gateways reducer', function () {
  const defaultState = {
    entities: {},
    selectedGateway: null,
    statistics: {},
  }

  it('should return the initial state', function () {
    expect(reducer(undefined, { type: '@@TEST_INIT', payload: {}})).toEqual(defaultState)
  })

  it('should ignore `getGatewayFailure` action', function () {
    expect(reducer(defaultState, getGatewayFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('should ignore `updateGatewayFailure` action', function () {
    expect(reducer(defaultState, updateGatewayFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('should ignore `getGatewaysListFailure` action', function () {
    expect(reducer(defaultState, getGatewaysListFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('should ignore `updateGateway` action', function () {
    expect(reducer(defaultState, updateGateway('test-id', {}))).toEqual(defaultState)
  })

  it('should ignore `deleteGatewayFailure` action', function () {
    expect(reducer(defaultState, deleteGatewayFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('should ignore `deleteGateway` action', function () {
    expect(reducer(defaultState, deleteGateway('test-id'))).toEqual(defaultState)
  })

  describe('requests single gateway', function () {
    const testGatewayId = 'tesrt-gtw-id'
    const testGateway = { ids: { gateway_id: testGatewayId }, name: 'test-gtw-name' }
    let newState

    beforeAll(function () {
      newState = reducer(defaultState, getGateway(testGatewayId))
    })

    it('should set `selectedGateway` on `getGateway` action', function () {
      expect(newState.selectedGateway).toEqual(testGatewayId)
    })

    it('should not update `entities` on `getGateway` action', function () {
      expect(newState.entities).toEqual(defaultState.entities)
    })

    describe('receives gateway', function () {
      beforeAll(function () {
        newState = reducer(newState, getGatewaySuccess(testGateway))
      })

      it('should not change `selectedGateway` on `getGatewaySuccess` action', function () {
        expect(newState.selectedGateway).toEqual(testGatewayId)
      })

      it('should add new gateway to `entities` on `getGatewaySuccess` action', function () {
        expect(Object.keys(newState.entities)).toHaveLength(1)
        expect(newState.entities[testGatewayId]).toEqual(testGateway)
      })

      describe('updates gateway', function () {
        const updatedTestGateway = { ids: { gateway_id: testGatewayId }, name: 'updated-test-gtw' }
        let updatedState

        beforeAll(function () {
          updatedState = reducer(newState, updateGatewaySuccess(updatedTestGateway))
        })

        it('should not change `selectedGateway` on `updateGatewaySuccess` action', function () {
          expect(updatedState.selectedGateway).toEqual(testGatewayId)
        })

        it('should update gateway in `entities` on `updateGatewaySuccess` action', function () {
          expect(updatedState.entities[testGatewayId].name).toEqual(updatedTestGateway.name)
        })
      })

      describe('deletes gateway', function () {
        let updatedState

        beforeAll(function () {
          updatedState = reducer(newState, deleteGatewaySuccess({ id: testGatewayId }))
        })

        it('should remove `selectedGateway` on `deleteGatewaySuccess` action', function () {
          expect(updatedState.selectedGateway).toBeNull()
        })

        it('should remove gateway in `entities` on `deleteGatewaySuccess` action', function () {
          expect(updatedState.entities[testGatewayId]).toBeUndefined()
        })
      })

      describe('requests another gateway', function () {
        const otherTestGatewayId = 'another-test-gtw-id'
        const otherTestGateway = { ids: { gateway_id: otherTestGatewayId }, name: 'another-test-gtw' }
        let updatedState

        beforeAll(function () {
          updatedState = reducer(newState, getGateway(otherTestGatewayId))
        })

        it('should set `selectedGateway` on `getGateway` action', function () {
          expect(updatedState.selectedGateway).toEqual(otherTestGatewayId)
        })

        it('should not update `entities` on `getGateway` action', function () {
          expect(Object.keys(updatedState.entities)).toHaveLength(1)
          expect(updatedState.entities[testGatewayId]).toEqual(testGateway)
        })

        describe('receives gateway', function () {
          beforeAll(function () {
            updatedState = reducer(updatedState, getGatewaySuccess(otherTestGateway))
          })

          it('should not change `selectedGateway` on `getGatewaySuccess` action', function () {
            expect(updatedState.selectedGateway).toEqual(otherTestGatewayId)
          })

          it('should keep previously received gateway in `entities`', function () {
            expect(updatedState.entities[testGatewayId]).toEqual(testGateway)
          })

          it('should add new gateway to `entities` on `getGatewaySuccess`', function () {
            expect(Object.keys(updatedState.entities)).toHaveLength(2)
            expect(updatedState.entities[otherTestGatewayId]).toEqual(otherTestGateway)
          })
        })
      })

      describe('requests a list of gateways', function () {
        beforeAll(function () {
          newState = reducer(newState, getGatewaysList({}))
        })

        it('should not change `selectedGateway` on `getGatewaysList` action', function () {
          expect(newState.selectedGateway).toEqual(testGatewayId)
        })

        it('should not change `entities` on `getGatewaysList` action', function () {
          expect(Object.keys(newState.entities)).toHaveLength(1)
          expect(newState.entities[testGatewayId]).toEqual(testGateway)
        })

        describe('receives a list of gateways', function () {
          const entities = [
            { ids: { gateway_id: 'test-gtw-1' }, name: 'test-gtw-1' },
            { ids: { gateway_id: 'test-gtw-2' }, name: 'test-gtw-2' },
            { ids: { gateway_id: 'test-gtw-3' }, name: 'test-gtw-3' },
          ]
          const totalCount = entities.length

          beforeAll(function () {
            newState = reducer(newState, getGatewaysListSuccess({ entities, totalCount }))
          })

          it('should not remove previously received gateway on `getGatewaysListSuccess` action', function () {
            expect(newState.entities[testGatewayId]).toEqual(testGateway)
          })

          it('should add new gateways to `entities` on `getGatewaysListSuccess`', function () {
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
