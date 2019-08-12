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

import getTableDriver from './index_driver'
import { noDataProps, paginatedProps, sortableProps, clickableProps } from './test-data'

describe('Table', function() {
  let driver = null

  beforeEach(function() {
    driver = getTableDriver()
  })

  describe('has no data provided', function() {
    beforeEach(function() {
      driver.when.created(noDataProps)
    })

    it('should match snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    it('should include the empty message', function() {
      expect(driver.is.empty()).toBeTruthy()
    })

    it('should have 0 data cells', function() {
      expect(driver.get.dataCellsCount()).toBe(0)
    })

    it('should have 2 head cells', function() {
      expect(driver.get.headCellsCount()).toBe(2)
    })
  })

  describe('has clickable rows', function() {
    let onRowClick = null

    beforeEach(function() {
      onRowClick = jest.fn()
      driver.when.created({
        ...clickableProps,
        onRowClick,
      })
    })

    it('should match snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    it('should have 2 head cells', function() {
      expect(driver.get.headCellsCount()).toBe(2)
    })

    it('should have 4 data cells', function() {
      expect(driver.get.dataCellsCount()).toBe(4)
    })

    it('the `onRowClick` function should no be called', function() {
      expect(onRowClick.mock.calls).toHaveLength(0)
    })

    describe('the user clicks the first row', function() {
      beforeEach(function() {
        driver.when.rowClicked(0)
      })

      it('the `onRowClick` function should be called once', function() {
        expect(onRowClick.mock.calls).toHaveLength(1)
      })

      describe('the user clicks the second row', function() {
        beforeEach(function() {
          driver.when.rowClicked(1)
        })

        it('the `onRowClick` function should be called twice', function() {
          expect(onRowClick.mock.calls).toHaveLength(2)
        })
      })
    })
  })

  describe('is paginated', function() {
    beforeEach(function() {
      driver.when.created(paginatedProps)
    })

    it('should match snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    it('should display the pagination', function() {
      expect(driver.is.paginated()).toBeTruthy()
    })
  })

  describe('is sortable', function() {
    let onSortRequest = null

    beforeEach(function() {
      onSortRequest = jest.fn()
      driver.when.created({
        ...sortableProps,
        onSortRequest,
      })
    })

    it('should match snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    it('should have 3 head cells', function() {
      expect(driver.get.headCellsCount()).toBe(3)
    })

    it('should have 2 sort buttons', function() {
      expect(driver.get.sortButtonsCount()).toBe(2)
    })

    it('should have 6 data cells', function() {
      expect(driver.get.dataCellsCount()).toBe(6)
    })

    it('should not have active sort buttons', function() {
      expect(driver.is.sortButtonActive(0)).toBeFalsy()
      expect(driver.is.sortButtonActive(1)).toBeFalsy()
    })

    it('the `onSortRequest` function should not be called', function() {
      expect(onSortRequest.mock.calls).toHaveLength(0)
    })

    describe('the users clicks on the first sort button', function() {
      beforeEach(function() {
        driver.when.sortButtonPressed(0)
        driver.when.updated({ order: 'asc', orderBy: 'test-column-name' })
      })

      it('the button should become active', function() {
        expect(driver.is.sortButtonActive(0)).toBeTruthy()
      })

      it('the `onSortRequest` function should be called once', function() {
        expect(onSortRequest.mock.calls).toHaveLength(1)
      })

      it('the `onSortRequest` function should be called with correct arguments', function() {
        const columnName = sortableProps.headers[0].name
        const order = 'asc'

        expect(onSortRequest.mock.calls[0][0]).toBe(order)
        expect(onSortRequest.mock.calls[0][1]).toBe(columnName)
      })

      describe('the user clicks on the second sort button', function() {
        beforeEach(function() {
          driver.when.sortButtonPressed(1)
          driver.when.updated({ order: 'asc', orderBy: 'test-column-name3' })
        })

        it('the first button should no longer be active', function() {
          expect(driver.is.sortButtonActive(0)).toBeFalsy()
        })

        it('the second button should become active', function() {
          expect(driver.is.sortButtonActive(1)).toBeTruthy()
        })

        it('the `onSortRequest` function should be called twice', function() {
          expect(onSortRequest.mock.calls).toHaveLength(2)
        })

        it('the `onSortRequest` function should be called with correct arguments', function() {
          const columnName = sortableProps.headers[2].name
          const order = 'asc'

          expect(onSortRequest.mock.calls[1][0]).toBe(order)
          expect(onSortRequest.mock.calls[1][1]).toBe(columnName)
        })
      })

      describe('the user clicks on the first sort button for the second time', function() {
        beforeEach(function() {
          driver.when.sortButtonPressed(0)
          driver.when.updated({ order: 'desc', orderBy: 'test-column-name' })
        })

        it('the button should stay active', function() {
          expect(driver.is.sortButtonActive(0)).toBeTruthy()
        })

        it('the second sort button should not be active', function() {
          expect(driver.is.sortButtonActive(1)).toBeFalsy()
        })

        it('the `onSortRequest` function should be called twice', function() {
          expect(onSortRequest.mock.calls).toHaveLength(2)
        })

        it('the `onSortRequest` function should be called with correct arguments', function() {
          const columnName = sortableProps.headers[0].name
          const order = 'desc'

          expect(onSortRequest.mock.calls[1][0]).toBe(order)
          expect(onSortRequest.mock.calls[1][1]).toBe(columnName)
        })

        describe('the user clicks on the first sort button for the third time', function() {
          beforeEach(function() {
            driver.when.sortButtonPressed(0)
            driver.when.updated({ order: undefined, orderBy: undefined })
          })

          it('the button should no longer be active', function() {
            expect(driver.is.sortButtonActive(0)).toBeFalsy()
          })

          it('the second sort button should not be active', function() {
            expect(driver.is.sortButtonActive(1)).toBeFalsy()
          })

          it('the `onSortRequest` function should be called three times', function() {
            expect(onSortRequest.mock.calls).toHaveLength(3)
          })

          it('the `onSortRequest` function should be called with correct arguments', function() {
            const columnName = undefined
            const order = undefined

            expect(onSortRequest.mock.calls[2][0]).toBe(order)
            expect(onSortRequest.mock.calls[2][1]).toBe(columnName)
          })
        })
      })
    })
  })
})
