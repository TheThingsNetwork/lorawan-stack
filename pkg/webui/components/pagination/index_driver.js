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

import Pagination from '.'

export default function () {
  const driver = {
    component: undefined,
    when: {
      created (props) {
        const wrapper = shallow(
          <Pagination {...props} />
        )

        if (wrapper.type() !== null) {
          driver.component = wrapper.dive()
          return driver
        }

        return undefined
      },
      navigatedNextPage () {
        driver.get.nextNavigationButton().simulate('click', {
          preventDefault: () => undefined,
        })
      },
      navigatedLastPage () {
        driver.get.lastPage().simulate('click', {
          preventDefault: () => undefined,
        })
      },
    },
    is: {
      pageSelected (page) {
        return driver.get.page(page).props().selected
      },
      firstPageSelected () {
        return driver.is.pageSelected(0)
      },
      lastPageSelected () {
        return driver.is.pageSelected(driver.get.pages().length - 1)
      },
      prevNavigationDisabled () {
        return driver.get.prevNavigation().hasClass('itemDisabled')
      },
      nextNavigationDisabled () {
        return driver.get.nextNavigation().hasClass('itemDisabled')
      },
    },
    get: {
      pages () {
        return driver.component.find('PageView')
      },
      page (page) {
        return driver.get.pages().at(page)
      },
      lastPage () {
        return driver.get.page(driver.get.pages().length - 1)
      },
      prevNavigation () {
        return driver.component.find('li').first()
      },
      nextNavigation () {
        return driver.component.find('li').last()
      },
      prevNavigationButton () {
        return driver.get.prevNavigation().find('a').first()
      },
      nextNavigationButton () {
        return driver.get.nextNavigation().find('a').first()
      },
    },
  }

  return driver
}
