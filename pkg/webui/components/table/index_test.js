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

describe('<Table />', function() {
  let driver = null

  beforeEach(function() {
    driver = getTableDriver()
  })

  describe('when it has no data provided', function() {
    beforeEach(function() {
      driver.when.created(noDataProps)
    })

    it('matches snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    it('includes the empty message', function() {
      expect(driver.is.empty()).toBeTruthy()
    })

    it('has 0 data cells', function() {
      expect(driver.get.dataCellsCount()).toBe(0)
    })

    it('has 2 head cells', function() {
      expect(driver.get.headCellsCount()).toBe(2)
    })
  })

  describe('when it has clickable rows', function() {
    let onRowClick = null

    beforeEach(function() {
      onRowClick = jest.fn()
      driver.when.created({
        ...clickableProps,
        onRowClick,
      })
    })

    it('matches snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    it('has 2 head cells', function() {
      expect(driver.get.headCellsCount()).toBe(2)
    })

    it('has 4 data cells', function() {
      expect(driver.get.dataCellsCount()).toBe(4)
    })

    it('does not call `onRowClick` function', function() {
      expect(onRowClick.mock.calls).toHaveLength(0)
    })

    describe('when the user clicks the first row', function() {
      beforeEach(function() {
        driver.when.rowClicked(0)
      })

      it('calls the `onRowClick` function once', function() {
        expect(onRowClick.mock.calls).toHaveLength(1)
      })

      describe('when the user clicks the second row', function() {
        beforeEach(function() {
          driver.when.rowClicked(1)
        })

        it('calls the `onRowClick` function twice', function() {
          expect(onRowClick.mock.calls).toHaveLength(2)
        })
      })
    })
  })

  describe('when it is paginated', function() {
    beforeEach(function() {
      driver.when.created(paginatedProps)
    })

    it('matches snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    it('displays the pagination', function() {
      expect(driver.is.paginated()).toBeTruthy()
    })
  })

  describe('when it is sortable', function() {
    let onSortRequest = null

    beforeEach(function() {
      onSortRequest = jest.fn()
      driver.when.created({
        ...sortableProps,
        onSortRequest,
      })
    })

    it('matches snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    it('has 3 head cells', function() {
      expect(driver.get.headCellsCount()).toBe(3)
    })

    it('has 2 sort buttons', function() {
      expect(driver.get.sortButtonsCount()).toBe(2)
    })

    it('has 6 data cells', function() {
      expect(driver.get.dataCellsCount()).toBe(6)
    })

    it('does not have active sort buttons', function() {
      expect(driver.is.sortButtonActive(0)).toBeFalsy()
      expect(driver.is.sortButtonActive(1)).toBeFalsy()
    })

    it('does not call `onSortRequest` function', function() {
      expect(onSortRequest.mock.calls).toHaveLength(0)
    })

    describe('when the user clicks on the first sort button', function() {
      beforeEach(function() {
        driver.when.sortButtonPressed(0)
        driver.when.updated({ order: 'asc', orderBy: 'test-column-name' })
      })

      it('activates the button', function() {
        expect(driver.is.sortButtonActive(0)).toBeTruthy()
      })

      it('calls the `onSortRequest` function once', function() {
        expect(onSortRequest.mock.calls).toHaveLength(1)
      })

      it('calls the `onSortRequest` function with correct arguments', function() {
        const columnName = sortableProps.headers[0].name
        const order = 'asc'

        expect(onSortRequest.mock.calls[0][0]).toBe(order)
        expect(onSortRequest.mock.calls[0][1]).toBe(columnName)
      })

      describe('when the user clicks on the second sort button', function() {
        beforeEach(function() {
          driver.when.sortButtonPressed(1)
          driver.when.updated({ order: 'asc', orderBy: 'test-column-name3' })
        })

        it('deactivates the first button', function() {
          expect(driver.is.sortButtonActive(0)).toBeFalsy()
        })

        it('activates the second button', function() {
          expect(driver.is.sortButtonActive(1)).toBeTruthy()
        })

        it('calls the `onSortRequest` function twice', function() {
          expect(onSortRequest.mock.calls).toHaveLength(2)
        })

        it('calls the `onSortRequest` function with correct arguments', function() {
          const columnName = sortableProps.headers[2].name
          const order = 'asc'

          expect(onSortRequest.mock.calls[1][0]).toBe(order)
          expect(onSortRequest.mock.calls[1][1]).toBe(columnName)
        })
      })

      describe('when the user clicks on the first sort button for the second time', function() {
        beforeEach(function() {
          driver.when.sortButtonPressed(0)
          driver.when.updated({ order: 'desc', orderBy: 'test-column-name' })
        })

        it('leaves the button activated', function() {
          expect(driver.is.sortButtonActive(0)).toBeTruthy()
        })

        it('deactivates the second button', function() {
          expect(driver.is.sortButtonActive(1)).toBeFalsy()
        })

        it('calls the `onSortRequest` function twice', function() {
          expect(onSortRequest.mock.calls).toHaveLength(2)
        })

        it('calls the `onSortRequest` function with correct arguments', function() {
          const columnName = sortableProps.headers[0].name
          const order = 'desc'

          expect(onSortRequest.mock.calls[1][0]).toBe(order)
          expect(onSortRequest.mock.calls[1][1]).toBe(columnName)
        })

        describe('when the user clicks on the first sort button for the third time', function() {
          beforeEach(function() {
            driver.when.sortButtonPressed(0)
            driver.when.updated({ order: undefined, orderBy: undefined })
          })

          it('deactivates the button', function() {
            expect(driver.is.sortButtonActive(0)).toBeFalsy()
          })

          it('deactivates the second button', function() {
            expect(driver.is.sortButtonActive(1)).toBeFalsy()
          })

          it('calls the `onSortRequest` function three times', function() {
            expect(onSortRequest.mock.calls).toHaveLength(3)
          })

          it('calls the `onSortRequest` function with correct arguments', function() {
            const columnName = 'test-column-name'

            expect(onSortRequest.mock.calls[0][0]).toBe('asc')
            expect(onSortRequest.mock.calls[1][0]).toBe('desc')
            expect(onSortRequest.mock.calls[2][0]).toBe('asc')
            expect(onSortRequest.mock.calls[2][1]).toBe(columnName)
          })
        })
      })
    })
  })
})
