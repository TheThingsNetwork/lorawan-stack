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

describe('applications reducer', function() {
  const defaultState = {
    entities: {},
    selectedApplication: null,
  }

  it('should return the initial state', function() {
    expect(reducer(undefined, { type: '@@TEST_INIT', payload: {} })).toEqual(defaultState)
  })

  it('should ignore `getApplicationFailure` action', function() {
    expect(reducer(defaultState, getApplicationFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('should ignore `updateApplicationFailure` action', function() {
    expect(reducer(defaultState, updateApplicationFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('should ignore `getApplicationsFailure` action', function() {
    expect(reducer(defaultState, getApplicationsFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('should ignore `updateApplication` action', function() {
    expect(reducer(defaultState, updateApplication('test-id', {}))).toEqual(defaultState)
  })

  it('should ignore `deleteApplicationFailure` action', function() {
    expect(reducer(defaultState, deleteApplicationFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('should ignore `deleteApplication` action', function() {
    expect(reducer(defaultState, deleteApplication('test-id'))).toEqual(defaultState)
  })

  describe('requests single application', function() {
    const testApplicationId = 'test-app-id'
    const testApplication = { ids: { application_id: testApplicationId }, name: 'test-app' }
    let newState

    beforeAll(function() {
      newState = reducer(defaultState, getApplication(testApplicationId))
    })

    it('should set `selectedApplication` on `getApplication` action', function() {
      expect(newState.selectedApplication).toEqual(testApplicationId)
    })

    it('should not update `entities` on `getApplication` action', function() {
      expect(newState.entities).toEqual(defaultState.entities)
    })

    describe('receives application', function() {
      beforeAll(function() {
        newState = reducer(newState, getApplicationSuccess(testApplication))
      })

      it('should not change `selectedApplication` on `getApplicationSuccess`', function() {
        expect(newState.selectedApplication).toEqual(testApplicationId)
      })

      it('should add new application to `entities` on `getApplicationSuccess`', function() {
        expect(Object.keys(newState.entities)).toHaveLength(1)
        expect(newState.entities[testApplicationId]).toEqual(testApplication)
      })

      describe('updates application', function() {
        const updatedTestApplication = {
          ids: { application_id: testApplicationId },
          name: 'updated-test-app',
        }
        let updatedState

        beforeAll(function() {
          updatedState = reducer(newState, updateApplicationSuccess(updatedTestApplication))
        })

        it('should not change `selectedApplication` on `updateApplicationSuccess`', function() {
          expect(updatedState.selectedApplication).toEqual(testApplicationId)
        })

        it('should update application in `entities` on `updateApplicationSuccess` action', function() {
          expect(updatedState.entities[testApplicationId].name).toEqual(updatedTestApplication.name)
        })
      })

      describe('deletes application', function() {
        let updatedState

        beforeAll(function() {
          updatedState = reducer(newState, deleteApplicationSuccess({ id: testApplicationId }))
        })

        it('should remove `selectedApplication` on `deleteApplicationSuccess`', function() {
          expect(updatedState.selectedApplication).toBeNull()
        })

        it('should remove application in `entities` on `deleteApplicationSuccess` action', function() {
          expect(updatedState.entities[testApplicationId]).toBeUndefined()
        })
      })

      describe('requests another application', function() {
        const otherTestApplicationId = 'another-test-app-id'
        const otherTestApplication = {
          ids: { application_id: otherTestApplicationId },
          name: 'test-app',
        }
        let updatedState

        beforeAll(function() {
          updatedState = reducer(newState, getApplication(otherTestApplicationId))
        })

        it('should set `selectedApplication` on `getApplication` action', function() {
          expect(updatedState.selectedApplication).toEqual(otherTestApplicationId)
        })

        it('should not update `entities` on `getApplication` action', function() {
          expect(Object.keys(updatedState.entities)).toHaveLength(1)
          expect(updatedState.entities[testApplicationId]).toEqual(testApplication)
        })

        describe('receives application', function() {
          beforeAll(function() {
            updatedState = reducer(updatedState, getApplicationSuccess(otherTestApplication))
          })

          it('should not change `selectedApplication` on `getApplicationSuccess`', function() {
            expect(updatedState.selectedApplication).toEqual(otherTestApplicationId)
          })

          it('should keep previously received application in `entities`', function() {
            expect(updatedState.entities[testApplicationId]).toEqual(testApplication)
          })

          it('should add new application to `entities` on `getApplicationSuccess`', function() {
            expect(Object.keys(updatedState.entities)).toHaveLength(2)
            expect(updatedState.entities[otherTestApplicationId]).toEqual(otherTestApplication)
          })
        })
      })

      describe('requests a list of applications', function() {
        beforeAll(function() {
          newState = reducer(newState, getApplicationsList({}))
        })

        it('should not change `selectedApplication` on `getApplicationList` action', function() {
          expect(newState.selectedApplication).toEqual(testApplicationId)
        })

        it('should not change `entities` on `getApplicationList` action', function() {
          expect(Object.keys(newState.entities)).toHaveLength(1)
          expect(newState.entities[testApplicationId]).toEqual(testApplication)
        })

        describe('receives a list of applications', function() {
          const entities = [
            { ids: { application_id: 'test-app-1' }, name: 'test-app-1' },
            { ids: { application_id: 'test-app-2' }, name: 'test-app-2' },
            { ids: { application_id: 'test-app-3' }, name: 'test-app-3' },
          ]
          const totalCount = entities.length

          beforeAll(function() {
            newState = reducer(newState, getApplicationsSuccess({ entities, totalCount }))
          })

          it('should not remove previously received application on `getApplicationsSuccess` action', function() {
            expect(newState.entities[testApplicationId]).toEqual(testApplication)
          })

          it('should add new applications to `entities` on `getApplicationSuccess`', function() {
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
