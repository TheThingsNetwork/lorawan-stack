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

import React from 'react'

import getSideNavigationDriver from './index_driver'
import SideNavigation from '.'

describe('SideNavigation', function() {
  let driver = null

  beforeEach(function() {
    driver = getSideNavigationDriver()
  })

  describe('is flat', function() {
    const props = {
      header: {
        title: 'test-header-title',
        icon: 'application',
      },
      children: (
        <React.Fragment>
          <SideNavigation.Item title="test-tile" path="/test-path" />
          <SideNavigation.Item title="test-tile2" path="/test-path2" />
        </React.Fragment>
      ),
    }

    beforeEach(function() {
      driver.when.created(props)
    })

    it('should match snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    it('should render correct number of items', function() {
      expect(driver.get.itemsCount()).toBe(2)
    })
    it('should have no collapsable items', function() {
      expect(driver.get.collapsableItemsCount()).toBe(0)
    })

    describe('the user minimizes the side navigation', function() {
      beforeEach(function() {
        driver.when.minimized()
      })

      it('should become minimized', function() {
        expect(driver.is.minimized()).toBeTruthy()
      })

      describe('the user expands the side navigation back to normal state', function() {
        beforeEach(function() {
          driver.when.minimized()
        })

        it('should not be minimized', function() {
          expect(driver.is.minimized()).toBeFalsy()
        })
      })
    })
  })

  describe('is nested', function() {
    const props = {
      header: {
        title: 'test-header-title',
        icon: 'application',
      },
      children: (
        <React.Fragment>
          <SideNavigation.Item title="test-title" path="/test-path">
            <SideNavigation.Item title="nested-test-tile" path="/nested-test-title" />
          </SideNavigation.Item>
          <SideNavigation.Item title="test-title2">
            <SideNavigation.Item title="nested-test-tile2" path="/nested-test-title2" />
            <SideNavigation.Item title="nested-test-tile3" path="/nested-test-title3" />
          </SideNavigation.Item>
          <SideNavigation.Item title="test-title3" path="/test-title3" />
        </React.Fragment>
      ),
    }

    beforeEach(function() {
      driver.when.created(props)
    })

    it('should match snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    it('should render correct number of top-level items', function() {
      expect(driver.get.itemsCount()).toBe(3)
    })

    it('should have correct number of collapsable items', function() {
      expect(driver.get.collapsableItemsCount()).toBe(2)
    })

    describe('the user minimizes the side navigation', function() {
      beforeEach(function() {
        driver.when.minimized()
      })

      it('should become minimized', function() {
        expect(driver.is.minimized()).toBeTruthy()
      })

      describe('the user expands the side navigation back to normal state', function() {
        beforeEach(function() {
          driver.when.minimized()
        })

        it('should not be minimized', function() {
          expect(driver.is.minimized()).toBeFalsy()
        })
      })
    })
  })
})
