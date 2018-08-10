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

import getHeaderCellDriver from './index_driver'

describe('HeaderCell', function () {
  let driver = null

  beforeEach(function () {
    driver = getHeaderCellDriver()
  })

  describe('is not sortable', function () {
    beforeEach(function () {
      driver.when.created({
        name: 'id',
        content: 'Id',
        sortable: false,
      })
    })

    it('should match snapshot', function () {
      expect(driver.component).toMatchSnapshot()
    })
  })

  describe('is sortable', function () {
    let onSortSpy = null
    const props = {
      name: 'id',
      content: 'Id',
      sortable: true,
    }

    beforeEach(function () {
      onSortSpy = jest.fn()
      driver.when.created({
        ...props,
        onSort: onSortSpy,
      })
    })

    it('should match snapshot', function () {
      expect(driver.component).toMatchSnapshot()
    })

    describe('the user presses on the sort button', function () {
      beforeEach(function () {
        driver.when.sortButtonPressed()
      })

      it('should call `onSort` with the cell name', function () {
        expect(onSortSpy.mock.calls[0][0]).toBe(props.name)
      })

      it('should call `onSort` only once', function () {
        expect(onSortSpy.mock.calls).toHaveLength(1)
      })
    })
  })

  describe('is active', function () {
    const props = {
      name: 'id',
      content: 'Id',
      sortable: true,
      active: true,
    }

    beforeEach(function () {
      driver.when.created(props)
    })

    it('should match snapshot', function () {
      expect(driver.component).toMatchSnapshot()
    })
  })
})
