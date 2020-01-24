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

import getSideNavigationItemDriver from './index_driver'
import { SideNavigationItem } from '.'

describe('SideNavigationItem', function() {
  let driver = null
  const location = {
    pathname: '/',
  }

  beforeEach(function() {
    driver = getSideNavigationItemDriver()
  })

  describe('is flat', function() {
    beforeEach(function() {
      driver.when.created({
        title: 'test-title',
        path: '/test-title',
        depth: 0,
        isActive: false,
      })
    })

    it('should match snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })
  })

  describe('is collapsable', function() {
    beforeEach(function() {
      driver.when.created({
        title: 'test-title',
        depth: 0,
        isActive: false,
        children: (
          <React.Fragment>
            <SideNavigationItem location={location} title="nested-title" path="/nested-title" />
            <SideNavigationItem location={location} title="nested-title2" path="/nested-title2" />
          </React.Fragment>
        ),
      })
    })

    it('should match snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    describe('the user selects the item', function() {
      beforeEach(function() {
        driver.when.itemSelected()
      })

      it('should toggle the isExpanded state to true', function() {
        expect(driver.component.state('isExpanded')).toBe(true)
      })

      it('should toggle the isExpanded state to false again', function() {
        driver.when.itemSelected()
        expect(driver.component.state('isExpanded')).toBe(false)
      })
    })
  })
})
