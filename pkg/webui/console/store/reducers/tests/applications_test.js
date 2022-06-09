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

describe('Applications reducer', () => {
  const defaultState = {
    entities: {},
    derived: {},
    selectedApplication: null,
    applicationDeviceCounts: {},
  }

  it('returns the initial state', () => {
    expect(reducer(undefined, { type: '@@TEST_INIT', payload: {} })).toEqual(defaultState)
  })

  it('ignores `getApplicationFailure` action', () => {
    expect(reducer(defaultState, getApplicationFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('ignores `updateApplicationFailure` action', () => {
    expect(reducer(defaultState, updateApplicationFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('ignores `getApplicationsFailure` action', () => {
    expect(reducer(defaultState, getApplicationsFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('ignores `updateApplication` action', () => {
    expect(reducer(defaultState, updateApplication('test-id', {}))).toEqual(defaultState)
  })

  it('ignores `deleteApplicationFailure` action', () => {
    expect(reducer(defaultState, deleteApplicationFailure({ status: 404 }))).toEqual(defaultState)
  })

  it('ignores `deleteApplication` action', () => {
    expect(reducer(defaultState, deleteApplication('test-id'))).toEqual(defaultState)
  })

  describe('when requesting a single application', () => {
    const testApplicationId = 'test-app-id'
    const testApplication = { ids: { application_id: testApplicationId }, name: 'test-app' }
    let newState

    beforeAll(() => {
      newState = reducer(defaultState, getApplication(testApplicationId))
    })

    it('sets `selectedApplication` on `getApplication` action', () => {
      expect(newState.selectedApplication).toEqual(testApplicationId)
    })

    it('does not update `entities` on `getApplication` action', () => {
      expect(newState.entities).toEqual(defaultState.entities)
    })

    describe('when receiving an application', () => {
      beforeAll(() => {
        newState = reducer(newState, getApplicationSuccess(testApplication))
      })

      it('does change `selectedApplication` on `getApplicationSuccess`', () => {
        expect(newState.selectedApplication).toEqual(testApplicationId)
      })

      it('adds new application to `entities` on `getApplicationSuccess`', () => {
        expect(Object.keys(newState.entities)).toHaveLength(1)
        expect(newState.entities[testApplicationId]).toEqual(testApplication)
      })

      describe('when it updates application', () => {
        const updatedTestApplication = {
          ids: { application_id: testApplicationId },
          name: 'updated-test-app',
        }
        let updatedState

        beforeAll(() => {
          updatedState = reducer(newState, updateApplicationSuccess(updatedTestApplication))
        })

        it('does not change `selectedApplication` on `updateApplicationSuccess`', () => {
          expect(updatedState.selectedApplication).toEqual(testApplicationId)
        })

        it('updates application in `entities` on `updateApplicationSuccess` action', () => {
          expect(updatedState.entities[testApplicationId].name).toEqual(updatedTestApplication.name)
        })
      })

      describe('when deleting an application', () => {
        let updatedState

        beforeAll(() => {
          updatedState = reducer(newState, deleteApplicationSuccess({ id: testApplicationId }))
        })

        it('removes `selectedApplication` on `deleteApplicationSuccess`', () => {
          expect(updatedState.selectedApplication).toBeNull()
        })

        it('removes application in `entities` on `deleteApplicationSuccess` action', () => {
          expect(updatedState.entities[testApplicationId]).toBeUndefined()
        })
      })

      describe('when requesting another application', () => {
        const otherTestApplicationId = 'another-test-app-id'
        const otherTestApplication = {
          ids: { application_id: otherTestApplicationId },
          name: 'test-app',
        }
        let updatedState

        beforeAll(() => {
          updatedState = reducer(newState, getApplication(otherTestApplicationId))
        })

        it('sets `selectedApplication` on `getApplication` action', () => {
          expect(updatedState.selectedApplication).toEqual(otherTestApplicationId)
        })

        it('does not update `entities` on `getApplication` action', () => {
          expect(Object.keys(updatedState.entities)).toHaveLength(1)
          expect(updatedState.entities[testApplicationId]).toEqual(testApplication)
        })

        describe('when receiving application', () => {
          beforeAll(() => {
            updatedState = reducer(updatedState, getApplicationSuccess(otherTestApplication))
          })

          it('does not change `selectedApplication` on `getApplicationSuccess`', () => {
            expect(updatedState.selectedApplication).toEqual(otherTestApplicationId)
          })

          it('keeps previously received application in `entities`', () => {
            expect(updatedState.entities[testApplicationId]).toEqual(testApplication)
          })

          it('adds new application to `entities` on `getApplicationSuccess`', () => {
            expect(Object.keys(updatedState.entities)).toHaveLength(2)
            expect(updatedState.entities[otherTestApplicationId]).toEqual(otherTestApplication)
          })
        })
      })

      describe('when requesting a list of applications', () => {
        beforeAll(() => {
          newState = reducer(newState, getApplicationsList({}))
        })

        it('does not change `selectedApplication` on `getApplicationList` action', () => {
          expect(newState.selectedApplication).toEqual(testApplicationId)
        })

        it('does not change `entities` on `getApplicationList` action', () => {
          expect(Object.keys(newState.entities)).toHaveLength(1)
          expect(newState.entities[testApplicationId]).toEqual(testApplication)
        })

        describe('when receiving a list of applications', () => {
          const entities = [
            { ids: { application_id: 'test-app-1' }, name: 'test-app-1' },
            { ids: { application_id: 'test-app-2' }, name: 'test-app-2' },
            { ids: { application_id: 'test-app-3' }, name: 'test-app-3' },
          ]
          const totalCount = entities.length

          beforeAll(() => {
            newState = reducer(newState, getApplicationsSuccess({ entities, totalCount }))
          })

          it('does not remove previously received application on `getApplicationsSuccess` action', () => {
            expect(newState.entities[testApplicationId]).toEqual(testApplication)
          })

          it('adds new applications to `entities` on `getApplicationSuccess`', () => {
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
