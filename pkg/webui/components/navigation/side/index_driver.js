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

import { SideNavigation } from '.'

export default function() {
  const driver = {
    component: undefined,
    location: {
      pathname: '/',
    },
    when: {
      created(props) {
        driver.component = shallow(<SideNavigation location={location} {...props} />)

        return driver
      },
      minimized() {
        driver.get.hideButton().simulate('click')
        driver.component.update()
      },
    },
    is: {
      minimized() {
        return driver.component.state().isMinimized
      },
      expanded(index) {
        return !!driver.component.state().itemsExpanded[index].isOpen
      },
    },
    get: {
      list() {
        return driver.component.find('SideNavigationList').dive()
      },
      nestedList(index) {
        return driver.get
          .item(index)
          .dive()
          .find('CollapsableItem')
          .dive()
          .find('SideNavigationList')
      },
      hideButton() {
        return driver.component.find('[data-hook="side-nav-hide-button"]')
      },
      items() {
        return driver.get.list().children()
      },
      item(index) {
        return driver.get.items().at(index)
      },
      itemsCount() {
        return driver.get.items().length
      },
      collapsableItemsCount() {
        return driver.get.items().findWhere(i => i.props().children !== undefined).length
      },
      expandedItemsCount() {
        return driver.get.items().findWhere(i => i.props().isExpanded).length
      },
    },
  }

  return driver
}
