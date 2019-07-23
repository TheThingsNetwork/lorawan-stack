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

import {
  createNamedPaginationReducerById,
  createNamedPaginationReducer,
} from '../pagination'
import {
  createPaginationRequestActions,
  createPaginationByIdRequestActions,
  createPaginationDeleteActions,
  createPaginationByIdDeleteActions,
} from '../../actions/pagination'

describe('pagination reducers', function () {
  const NAME = 'ENTITY'
  const entityIdSelector = entity => entity.id

  describe('flat', function () {
    const reducer = createNamedPaginationReducer(NAME, entityIdSelector)
    const defaultState = { ids: [], totalCount: undefined }

    const {
      request,
      success,
      failure,
    } = createPaginationRequestActions(NAME)[1]
    const {
      request: requestDelete,
      success: successDelete,
      failure: failureDelete,
    } = createPaginationDeleteActions(NAME)[1]

    it('should return the initial state', function () {
      expect(reducer(undefined, { payload: {}})).toEqual(defaultState)
    })

    it('should ignore the get `request` action', function () {
      expect(reducer(defaultState, request())).toEqual(defaultState)
    })

    it('should ignore the get `failure` action', function () {
      expect(reducer(defaultState, failure())).toEqual(defaultState)
    })

    it('should ignore the delete `request` action', function () {
      expect(reducer(defaultState, requestDelete('test-id'))).toEqual(defaultState)
    })

    it('should ignore the delete `failure` action', function () {
      expect(reducer(defaultState, failureDelete())).toEqual(defaultState)
    })

    describe('receives the `success` action', function () {
      const entities = [
        { id: '1', name: 'name1' },
        { id: '2', name: 'name2' },
      ]
      const totalCount = entities.length
      const action = success({ entities, totalCount })

      let newState = null

      beforeAll(function () {
        newState = reducer(defaultState, action)
      })

      it('should update the state', function () {
        expect(newState).not.toEqual(defaultState)
      })

      it('should only store ids', function () {
        const { ids } = newState

        expect(ids).toEqual(entities.map(e => e.id))
      })

      it('should store `totalCount`', function () {
        const { totalCount: newTotalCount } = newState

        expect(newTotalCount).toEqual(totalCount)
      })

      describe('deletes an entity', function () {
        beforeAll(function () {
          newState = reducer(newState, successDelete({ id: entities[0].id }))
        })

        it('should decrease `totalCount`', function () {
          expect(newState.totalCount).toEqual(entities.length - 1)
        })

        it('should remove deleted id from `ids`', function () {
          expect(newState.ids).toHaveLength(entities.length - 1)
          expect(newState.ids).toEqual(
            expect.not.arrayContaining(entities)
          )
        })
      })
    })
  })

  describe('by id', function () {
    const reducer = createNamedPaginationReducerById(NAME, entityIdSelector)
    const defaultState = {}
    const entityId = 'parent-id'

    const {
      request,
      success,
      failure,
    } = createPaginationByIdRequestActions(NAME)[1]
    const {
      request: requestDelete,
      success: successDelete,
      failure: failureDelete,
    } = createPaginationByIdDeleteActions(NAME)[1]

    it('should return the initial state', function () {
      expect(reducer(undefined, { payload: {}})).toEqual(defaultState)
    })

    it('should ignore the `request` action', function () {
      expect(reducer(defaultState, request(entityId))).toEqual(defaultState)
    })

    it('should ignore the `failure` action', function () {
      expect(reducer(defaultState, failure({}))).toEqual(defaultState)
    })

    it('should ignore the delete `request` action', function () {
      expect(reducer(defaultState, requestDelete('test-id', 'target-test-id'))).toEqual(defaultState)
    })

    it('should ignore the delete `failure` action', function () {
      expect(reducer(defaultState, failureDelete('test-id'))).toEqual(defaultState)
    })

    describe('receives the `success` action', function () {
      const entities = [
        { id: '1', name: 'name1' },
        { id: '2', name: 'name2' },
      ]
      const totalCount = entities.length
      const action = success({ id: entityId, entities, totalCount })

      let newState = null

      beforeAll(function () {
        newState = reducer(defaultState, action)
      })

      it('should ignore without `id` in the payload', function () {
        const updatedState = reducer(defaultState, success({ entities: [], totalCount }))

        expect(updatedState).toEqual(defaultState)
      })

      it('should update the state', function () {
        expect(newState).not.toEqual(defaultState)
      })

      it('should store results per entity id', function () {
        const { [entityId]: results } = newState

        expect(results).not.toBeUndefined()
      })

      it('should only store ids', function () {
        const { [entityId]: results } = newState

        expect(results.ids).toEqual(entities.map(entityIdSelector))
      })

      it('should store `totalCount`', function () {
        const { [entityId]: results } = newState

        expect(results.totalCount).toEqual(totalCount)
      })

      describe('deletes entity', function () {
        let updatedState

        beforeAll(function () {
          updatedState = reducer(newState, successDelete({ id: entityId, targetId: entities[0].id }))
        })

        it('should decrease `totalCount`', function () {
          const { [entityId]: results } = updatedState
          expect(results.totalCount).toEqual(entities.length - 1)
        })

        it('should remove deleted id from `ids`', function () {
          const { [entityId]: results } = updatedState
          expect(results.ids).toHaveLength(entities.length - 1)
          expect(results.ids).toEqual(
            expect.not.arrayContaining(entities)
          )
        })
      })

      describe('receives the `success` action for another entity', function () {
        const otherEntityId = 'other-entity-id'
        const otherEntities = [
          { id: '3', name: 'name3' },
          { id: '4', name: 'name4' },
          { id: '5', name: 'name5' },
        ]
        const otherTotalCount = otherEntities.length
        const action = success({
          id: otherEntityId,
          entities: otherEntities,
          totalCount: otherTotalCount,
        })

        let otherNewState = null

        beforeAll(function () {
          otherNewState = reducer(newState, action)
        })

        it('should update the state', function () {
          expect(otherNewState).not.toEqual(newState)
        })

        it('should preserve previous entries', function () {
          const { [entityId]: results } = otherNewState

          expect(results).not.toBeUndefined()
          expect(results.ids).toEqual(entities.map(e => e.id))
          expect(results.totalCount).toEqual(totalCount)
        })

        it('should store results per entity id', function () {
          const { [otherEntityId]: results } = otherNewState

          expect(results).not.toBeUndefined()
        })

        it('should only store ids', function () {
          const { [otherEntityId]: results } = otherNewState

          expect(results.ids).toEqual(otherEntities.map(e => e.id))
        })

        it('should store `totalCount`', function () {
          const { [otherEntityId]: results } = otherNewState

          expect(results.totalCount).toEqual(otherTotalCount)
        })

        describe('deletes entity', function () {
          let updatedState

          beforeAll(function () {
            updatedState = reducer(otherNewState, successDelete({ id: otherEntityId, targetId: otherEntities[0].id }))
          })

          it('should decrease `totalCount`', function () {
            const { [entityId]: results } = updatedState
            expect(results.totalCount).toEqual(otherEntities.length - 1)
          })

          it('should remove deleted id from `ids`', function () {
            const { [entityId]: results } = updatedState
            expect(results.ids).toHaveLength(otherEntities.length - 1)
            expect(results.ids).toEqual(
              expect.not.arrayContaining(otherEntities)
            )
          })
        })
      })
    })
  })
})
