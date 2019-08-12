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

import { Cell, HeadCell, DataCell } from '.'

function getCellDriver(create) {
  const driver = {
    component: undefined,
    when: {
      created(props) {
        driver.component = create(props)
      },
    },
    get: {
      cell() {
        return driver.component.find(Cell).first()
      },
      cellRaw() {
        return driver.get.cell().dive()
      },
      cellType() {
        return driver.get.cellRaw().type()
      },
    },
  }

  return driver
}

export function getHeadCellDriver() {
  return getCellDriver(props => shallow(<HeadCell {...props} />))
}

export function getDataCellDriver() {
  return getCellDriver(props => shallow(<DataCell {...props} />))
}
