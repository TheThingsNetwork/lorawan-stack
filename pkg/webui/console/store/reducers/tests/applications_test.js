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

import reducer from '../applications'
import {
  getApplication,
  getApplicationSuccess,
  getApplicationFailure,
  getApplicationsList,
  getApplicationsSuccess,
  getApplicationsFailure,
  updateApplication,
  updateApplicationSuccess,
  updateApplicationFailure,
  deleteApplication,
  deleteApplicationSuccess,
  deleteApplicationFailure,
} from '../../actions/applications'

describe('Applications reducer', function () {
  const defaultState = {
    entities: {},
    selectedApplication: null,
  }

  it('returns the initial state', function () {
    expect(reducer(undefined, { type: '@@TEST_INIT', payload: {} })).toEqual(defaultState)
  })

  it('ignores `getApplicationFailure` action', function () {
    expect(reducer(defaultState, getApplicationFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('ignores `updateApplicationFailure` action', function () {
    expect(reducer(defaultState, updateApplicationFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('ignores `getApplicationsFailure` action', function () {
    expect(reducer(defaultState, getApplicationsFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('ignores `updateApplication` action', function () {
    expect(reducer(defaultState, updateApplication('test-id', {}))).toEqual(defaultState)
  })

  it('ignores `deleteApplicationFailure` action', function () {
    expect(reducer(defaultState, deleteApplicationFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('ignores `deleteApplication` action', function () {
    expect(reducer(defaultState, deleteApplication('test-id'))).toEqual(defaultState)
  })

  describe('when requesting a single application', function () {
    const testApplicationId = 'test-app-id'
    const testApplication = { ids: { application_id: testApplicationId }, name: 'test-app' }
    let newState

    beforeAll(function () {
      newState = reducer(defaultState, getApplication(testApplicationId))
    })

    it('sets `selectedApplication` on `getApplication` action', function () {
      expect(newState.selectedApplication).toEqual(testApplicationId)
    })

    it('does not update `entities` on `getApplication` action', function () {
      expect(newState.entities).toEqual(defaultState.entities)
    })

    describe('when receiving an application', function () {
      beforeAll(function () {
        newState = reducer(newState, getApplicationSuccess(testApplication))
      })

      it('does change `selectedApplication` on `getApplicationSuccess`', function () {
        expect(newState.selectedApplication).toEqual(testApplicationId)
      })

      it('adds new application to `entities` on `getApplicationSuccess`', function () {
        expect(Object.keys(newState.entities)).toHaveLength(1)
        expect(newState.entities[testApplicationId]).toEqual(testApplication)
      })

      describe('when it updates application', function () {
        const updatedTestApplication = {
          ids: { application_id: testApplicationId },
          name: 'updated-test-app',
        }
        let updatedState

        beforeAll(function () {
          updatedState = reducer(newState, updateApplicationSuccess(updatedTestApplication))
        })

        it('does not change `selectedApplication` on `updateApplicationSuccess`', function () {
          expect(updatedState.selectedApplication).toEqual(testApplicationId)
        })

        it('updates application in `entities` on `updateApplicationSuccess` action', function () {
          expect(updatedState.entities[testApplicationId].name).toEqual(updatedTestApplication.name)
        })
      })

      describe('when deleting an application', function () {
        let updatedState

        beforeAll(function () {
          updatedState = reducer(newState, deleteApplicationSuccess({ id: testApplicationId }))
        })

        it('removes `selectedApplication` on `deleteApplicationSuccess`', function () {
          expect(updatedState.selectedApplication).toBeNull()
        })

        it('removes application in `entities` on `deleteApplicationSuccess` action', function () {
          expect(updatedState.entities[testApplicationId]).toBeUndefined()
        })
      })

      describe('when requesting another application', function () {
        const otherTestApplicationId = 'another-test-app-id'
        const otherTestApplication = {
          ids: { application_id: otherTestApplicationId },
          name: 'test-app',
        }
        let updatedState

        beforeAll(function () {
          updatedState = reducer(newState, getApplication(otherTestApplicationId))
        })

        it('sets `selectedApplication` on `getApplication` action', function () {
          expect(updatedState.selectedApplication).toEqual(otherTestApplicationId)
        })

        it('does not update `entities` on `getApplication` action', function () {
          expect(Object.keys(updatedState.entities)).toHaveLength(1)
          expect(updatedState.entities[testApplicationId]).toEqual(testApplication)
        })

        describe('when receiving application', function () {
          beforeAll(function () {
            updatedState = reducer(updatedState, getApplicationSuccess(otherTestApplication))
          })

          it('does not change `selectedApplication` on `getApplicationSuccess`', function () {
            expect(updatedState.selectedApplication).toEqual(otherTestApplicationId)
          })

          it('keeps previously received application in `entities`', function () {
            expect(updatedState.entities[testApplicationId]).toEqual(testApplication)
          })

          it('adds new application to `entities` on `getApplicationSuccess`', function () {
            expect(Object.keys(updatedState.entities)).toHaveLength(2)
            expect(updatedState.entities[otherTestApplicationId]).toEqual(otherTestApplication)
          })
        })
      })

      describe('when requesting a list of applications', function () {
        beforeAll(function () {
          newState = reducer(newState, getApplicationsList({}))
        })

        it('does not change `selectedApplication` on `getApplicationList` action', function () {
          expect(newState.selectedApplication).toEqual(testApplicationId)
        })

        it('does not change `entities` on `getApplicationList` action', function () {
          expect(Object.keys(newState.entities)).toHaveLength(1)
          expect(newState.entities[testApplicationId]).toEqual(testApplication)
        })

        describe('when receiving a list of applications', function () {
          const entities = [
            { ids: { application_id: 'test-app-1' }, name: 'test-app-1' },
            { ids: { application_id: 'test-app-2' }, name: 'test-app-2' },
            { ids: { application_id: 'test-app-3' }, name: 'test-app-3' },
          ]
          const totalCount = entities.length

          beforeAll(function () {
            newState = reducer(newState, getApplicationsSuccess({ entities, totalCount }))
          })

          it('does not remove previously received application on `getApplicationsSuccess` action', function () {
            expect(newState.entities[testApplicationId]).toEqual(testApplication)
          })

          it('adds new applications to `entities` on `getApplicationSuccess`', function () {
            expect(Object.keys(newState.entities)).toHaveLength(4)
            for (const app of entities) {
              expect(newState.entities[app.ids.application_id]).toEqual(app)
            }
          })
        })
      })
    })
  })
})
