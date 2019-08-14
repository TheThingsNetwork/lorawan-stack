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
import Table from '.'

export default function() {
  const driver = {
    component: undefined,
    when: {
      created(props) {
        driver.component = shallow(<Table {...props} />)

        return driver
      },
      updated(props) {
        driver.component.setProps(props)
      },
      rowClicked(index) {
        driver.get.row(index + 1).simulate('click')
        driver.component.update()
      },
      sortButtonPressed(index) {
        driver.get
          .sortButton(index)
          .dive()
          .simulate('click')
        driver.component.update()
      },
    },
    is: {
      empty() {
        return driver.get.emptyMessage().exists()
      },
      paginated() {
        return driver.get.pagination().exists()
      },
      sortButtonActive(index) {
        return driver.get.sortButton(index).props().active
      },
    },
    get: {
      emptyMessage() {
        return driver.component.find('Empty').first()
      },
      headCellsCount() {
        return driver.get.headCells().length
      },
      headCells() {
        return driver.component.find('HeadCell')
      },
      dataCellsCount() {
        return driver.get.dataCells().length
      },
      dataCells() {
        return driver.component.find('DataCell')
      },
      sortButtonsCount() {
        return driver.get.sortButtons().length
      },
      sortButtons() {
        return driver.component.find('SortButton')
      },
      sortButton(index) {
        return driver.get.sortButtons().at(index)
      },
      pagination() {
        return driver.component.find('Pagination').first()
      },
      rows() {
        return driver.component.find('Row')
      },
      row(index) {
        return driver.get.rows().at(index)
      },
    },
  }

  return driver
}
