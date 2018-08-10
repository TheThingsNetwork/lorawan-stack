// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

describe('Table', function () {
  let driver = null

  beforeEach(function () {
    driver = getTableDriver()
  })

  describe('has not data', function () {
    beforeEach(function () {
      const props = {
        totalSize: 0,
        emptyMessage: 'No entries',
        onPageChange: () => 0,
        onSortByColumn: () => '',
        headers: [{
          name: 'id',
          displayName: 'Id',
        }, {
          name: 'desc',
          displayName: 'Description',
        }],
        rows: [],
      }

      driver.when.created(props)
    })

    it('should match snapshot', function () {
      expect(driver.component).toMatchSnapshot()
    })

    it('should display an empty table message', function () {
      expect(
        driver.is.emptyMessageShown()
      ).toBeTruthy()
    })
  })

  describe('has data', function () {
    const pageSize = 2
    const headers = [
      {
        name: 'id',
        displayName: 'Id',
        sortable: true,
      },
      {
        name: 'desc',
        displayName: 'Description',
      },
    ]

    describe('the number of rows is equal to the page size', function () {
      const data = [
        {
          id: 1,
          desc: 'Description1',
        },
        {
          id: 2,
          desc: 'Description2',
        },
      ]

      beforeEach(function () {
        driver.when.created({
          rows: data,
          headers,
          pageSize,
          onPageChange: () => 0,
          onSortByColumn: () => '',
          totalSize: data.length,
          emptyMessage: 'No entries',
        })
      })

      it('should match snapshot', function () {
        expect(driver.component).toMatchSnapshot()
      })

      it('should show all rows', function () {
        expect(driver.is.filledWithDataCells(data.length)).toBeTruthy()
      })
    })
  })
})
