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

// Mock the window inner width to ensure initial render in expanded state.
global.window.innerWidth = 1440

describe('<SideNavigation />', function() {
  let driver = null

  beforeEach(function() {
    driver = getSideNavigationDriver()
  })

  describe('when it is flat', function() {
    const props = {
      modifyAppContainerClasses: false,
      header: {
        title: 'test-header-title',
        icon: 'application',
        to: '/',
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

    it('matches snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    it('renders correct number of items', function() {
      expect(driver.get.itemsCount()).toBe(2)
    })
    it('has no collapsable items', function() {
      expect(driver.get.collapsableItemsCount()).toBe(0)
    })

    describe('when the user minimizes the side navigation', function() {
      beforeEach(function() {
        driver.when.minimized()
      })

      it('becomes minimized', function() {
        expect(driver.is.minimized()).toBeTruthy()
      })

      describe('when the user expands the side navigation back to normal state', function() {
        beforeEach(function() {
          driver.when.minimized()
        })

        it('becomes maximized again', function() {
          expect(driver.is.minimized()).toBeFalsy()
        })
      })
    })
  })

  describe('when it is nested', function() {
    const props = {
      modifyAppContainerClasses: false,
      header: {
        title: 'test-header-title',
        icon: 'application',
        to: '/',
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

    it('matches snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    it('renders correct number of top-level items', function() {
      expect(driver.get.itemsCount()).toBe(3)
    })

    it('has correct number of collapsable items', function() {
      expect(driver.get.collapsableItemsCount()).toBe(2)
    })

    describe('when the user minimizes the side navigation', function() {
      beforeEach(function() {
        driver.when.minimized()
      })

      it('becomes minimized', function() {
        expect(driver.is.minimized()).toBeTruthy()
      })

      describe('when the user expands the side navigation back to normal state', function() {
        beforeEach(function() {
          driver.when.minimized()
        })

        it('becomes maximized again', function() {
          expect(driver.is.minimized()).toBeFalsy()
        })
      })
    })
  })
})
