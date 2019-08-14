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

import getSideNavigationDriver from './index_driver'

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
      entries: [
        {
          title: 'test-title',
          path: '/test-path',
        },
        {
          title: 'test-title2',
          path: '/test-title2',
        },
      ],
    }

    beforeEach(function() {
      driver.when.created(props)
    })

    it('should match snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    it('should render correct number of items', function() {
      expect(driver.get.itemsCount()).toBe(props.entries.length)
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
      entries: [
        {
          title: 'test-title',
          nested: true,
          items: [
            {
              title: 'nested-test-title',
              path: '/nested-test-title',
            },
          ],
        },
        {
          title: 'test-title2',
          nested: true,
          items: [
            {
              title: 'nested-test-title2',
              path: '/nested-test-title2',
            },
            {
              title: 'nested-test-title3',
              path: '/nested-test-title3',
            },
          ],
        },
        {
          title: 'test-title3',
          path: '/test-title3',
        },
      ],
    }

    beforeEach(function() {
      driver.when.created(props)
    })

    it('should match snapshot', function() {
      expect(driver.component).toMatchSnapshot()
    })

    it('should render correct number of top-level items', function() {
      expect(driver.get.itemsCount()).toBe(props.entries.length)
    })

    it('should have correct number of collapsable items', function() {
      const count = props.entries.filter(i => i.nested).length
      expect(driver.get.collapsableItemsCount()).toBe(count)
    })

    describe('the user minimizes the side navigation', function() {
      beforeEach(function() {
        driver.when.minimized()
      })

      it('should become minimized', function() {
        expect(driver.is.minimized()).toBeTruthy()
      })

      describe('the user selects a top-level item', function() {
        const SELECTED_INDEX = 2

        beforeEach(function() {
          driver.when.linkSelected(SELECTED_INDEX)
        })

        it('should stay minimized', function() {
          expect(driver.is.minimized()).toBeTruthy()
        })
      })

      describe('the user selects the first closed and collapsable item', function() {
        const SELECTED_INDEX = 0

        beforeEach(function() {
          driver.when.itemSelected(SELECTED_INDEX)
        })

        it('should not be minimized', function() {
          expect(driver.is.minimized()).toBeFalsy()
        })

        it('should update the state', function() {
          expect(driver.is.expanded(SELECTED_INDEX)).toBeTruthy()
        })

        it('should expand the item', function() {
          expect(driver.is.itemExpanded(SELECTED_INDEX)).toBeTruthy()
        })

        it('should not be active', function() {
          expect(driver.is.itemActive(SELECTED_INDEX)).toBeFalsy()
        })

        describe('the user selects a link in the opened item', function() {
          const SELECTED_LINK_INDEX = 0

          beforeEach(function() {
            driver.when.nestedLinkSelected(SELECTED_INDEX, SELECTED_LINK_INDEX)
          })

          it('the item should become active', function() {
            expect(driver.is.itemActive(SELECTED_INDEX)).toBeTruthy()
          })
        })
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

    describe('the user selects a collapsable item', function() {
      const FST_SELECTED_INDEX = 0
      const SND_SELECTED_INDEX = 1

      beforeEach(function() {
        driver.when.itemSelected(FST_SELECTED_INDEX)
      })

      it('should update the state', function() {
        expect(driver.is.expanded(FST_SELECTED_INDEX)).toBeTruthy()
      })

      it('should expand the item', function() {
        expect(driver.is.itemExpanded(FST_SELECTED_INDEX)).toBeTruthy()
      })

      describe('the user selects the item again', function() {
        beforeEach(function() {
          driver.when.itemSelected(FST_SELECTED_INDEX)
        })

        it('should update the state', function() {
          expect(driver.is.expanded(FST_SELECTED_INDEX)).toBeFalsy()
        })

        it('should collapse the item', function() {
          expect(driver.is.itemExpanded(FST_SELECTED_INDEX)).toBeFalsy()
        })

        it('should have no expanded items', function() {
          expect(driver.get.expandedItemsCount()).toBe(0)
        })
      })

      describe('the user selects another collapsable item', function() {
        beforeEach(function() {
          driver.when.itemSelected(SND_SELECTED_INDEX)
        })

        it('should keep the first collapsable item in the state', function() {
          expect(driver.is.expanded(FST_SELECTED_INDEX)).toBeTruthy()
        })

        it('should keep the first collapsable item expanded', function() {
          expect(driver.is.itemExpanded(FST_SELECTED_INDEX)).toBeTruthy()
        })

        it('should update the state', function() {
          expect(driver.is.expanded(SND_SELECTED_INDEX)).toBeTruthy()
        })

        it('should expand the second item', function() {
          expect(driver.is.itemExpanded(SND_SELECTED_INDEX)).toBeTruthy()
        })

        it('should have `2` items expanded', function() {
          expect(driver.get.expandedItemsCount()).toBe(2)
        })
      })
    })
  })
})
