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

import React from 'react'
import Table from '.'

export default function () {
  const driver = {
    component: undefined,
    when: {
      created (props) {
        driver.component = shallow(
          <Table {...props} />
        )

        return driver
      },
    },
    is: {
      emptyMessageShown () {
        return driver.get.emptyMessage().exists()
      },
      filledWithDataCells (count) {
        return driver.get.dataCells().length === count
      },
    },
    get: {
      emptyMessage () {
        return driver.component.find('[data-hook="empty-message"]')
      },
      dataCells () {
        return driver.component.find('[data-hook="data-row"]')
      },
    },
  }

  return driver
}
