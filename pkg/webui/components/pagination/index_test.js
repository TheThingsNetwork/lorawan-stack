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

import getPaginationDriver from './index_driver'

describe('Pagination', function() {
  let driver = null

  beforeEach(function() {
    driver = getPaginationDriver()
  })

  describe('has only a single page', function() {
    describe('with hiding option disabled', function() {
      beforeEach(function() {
        driver.when.created({ pageCount: 1, hideIfOnlyOnePage: false })
      })
      it('should match snapshot', function() {
        expect(driver.component).toMatchSnapshot()
      })

      it('should disable the previous page navigation', function() {
        expect(driver.is.prevNavigationDisabled()).toBeTruthy()
      })

      it('should disable the next page navigation', function() {
        expect(driver.is.nextNavigationDisabled()).toBeTruthy()
      })
    })
    describe('with hiding option enabled', function() {
      beforeEach(function() {
        driver.when.created({ pageCount: 1 })
      })
      it('should not render', function() {
        expect(driver.component).toBe(undefined)
      })
    })
  })

  describe('has several pages', function() {
    beforeEach(function() {
      driver.when.created({ pageCount: 3 })
    })

    it('should match snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    it('should select the first page', function() {
      expect(driver.is.firstPageSelected()).toBeTruthy()
    })

    it('should disable the previous navigation', function() {
      expect(driver.is.prevNavigationDisabled()).toBeTruthy()
    })

    describe('the user moves to the next page', function() {
      beforeEach(function() {
        driver.when.navigatedNextPage()
      })

      it('should select the second page', function() {
        expect(driver.is.pageSelected(1)).toBeTruthy()
      })

      it('should enable the previous navigation', function() {
        expect(driver.is.prevNavigationDisabled()).toBeFalsy()
      })
    })

    describe('the user moves to the last page', function() {
      beforeEach(function() {
        driver.when.navigatedLastPage()
      })

      it('should select the last page', function() {
        expect(driver.is.lastPageSelected()).toBeTruthy()
      })

      it('should disable the the next navigation', function() {
        expect(driver.is.nextNavigationDisabled()).toBeTruthy()
      })
    })
  })
})
