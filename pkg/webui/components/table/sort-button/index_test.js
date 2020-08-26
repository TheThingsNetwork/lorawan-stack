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

import getSortButtonDriver from './index_driver'

describe('<SortButton />', function() {
  let driver = null
  let onSort = null

  beforeEach(function() {
    driver = getSortButtonDriver()
    onSort = jest.fn()
  })

  describe('when it is not active', function() {
    beforeEach(function() {
      driver.when.created({
        active: false,
        title: 'test-title',
        name: 'test-name',
        direction: undefined,
        onSort,
      })
    })

    it('matches snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    it('does not have the `active` style', function() {
      expect(driver.is.active()).toBeFalsy()
    })

    it('does not have styling for the `descending` style', function() {
      expect(driver.is.descending()).toBeFalsy()
    })

    describe('when the user presses the button', function() {
      beforeEach(function() {
        driver.when.buttonPressed()
      })

      it('calls the `onSort` function once', function() {
        expect(onSort.mock.calls).toHaveLength(1)
      })
    })
  })

  describe('when it is active', function() {
    describe('is in ascending direction', function() {
      beforeEach(function() {
        driver.when.created({
          active: true,
          title: 'test-title',
          name: 'test-name',
          direction: 'asc',
          onSort,
        })
      })

      it('matches snapshot', function() {
        expect(driver.component).toMatchSnapshot()
      })

      it('has the `active` style', function() {
        expect(driver.is.active()).toBeTruthy()
      })

      it('has styling for the `descending` style', function() {
        expect(driver.is.descending()).toBeFalsy()
      })

      describe('when the user presses the button', function() {
        beforeEach(function() {
          driver.when.buttonPressed()
        })

        it('calls the `onSort` function once', function() {
          expect(onSort.mock.calls).toHaveLength(1)
        })
      })
    })

    describe('when it is in descending direction', function() {
      beforeEach(function() {
        driver.when.created({
          active: true,
          title: 'test-title',
          name: 'test-name',
          direction: 'desc',
          onSort,
        })
      })

      it('matches snapshot', function() {
        expect(driver.component).toMatchSnapshot()
      })

      it('has the `active` style', function() {
        expect(driver.is.active()).toBeTruthy()
      })

      it('has styling for the `descending` style', function() {
        expect(driver.is.descending()).toBeTruthy()
      })

      describe('when the user presses the button', function() {
        beforeEach(function() {
          driver.when.buttonPressed()
        })

        it('calls the `onSort` function once', function() {
          expect(onSort.mock.calls).toHaveLength(1)
        })
      })
    })
  })
})
