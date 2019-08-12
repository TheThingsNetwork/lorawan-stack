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

import getSideNavigationItemDriver from './index_driver'

describe('SideNavigationItem', function() {
  let driver = null
  let onExpandSpy = null

  beforeEach(function() {
    driver = getSideNavigationItemDriver()
    onExpandSpy = jest.fn()
  })

  describe('is flat', function() {
    beforeEach(function() {
      driver.when.created({
        title: 'test-title',
        path: '/test-title',
        depth: 0,
        isCollapsable: false,
        isMinimized: false,
        isExpanded: false,
        isActive: false,
        items: [],
        onExpand: onExpandSpy,
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
        isCollapsable: true,
        isMinimized: false,
        isExpanded: true,
        isActive: false,
        onExpand: onExpandSpy,
        items: [
          {
            title: 'nested-title',
            path: '/nested-title',
          },
          {
            title: 'nested-title2',
            path: '/nested-title2',
          },
        ],
      })
    })

    it('should match snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    describe('the user selects the item', function() {
      beforeEach(function() {
        driver.when.itemSelected()
      })

      it('should call `onExpand` only once', function() {
        expect(onExpandSpy.mock.calls).toHaveLength(1)
      })
    })
  })
})
