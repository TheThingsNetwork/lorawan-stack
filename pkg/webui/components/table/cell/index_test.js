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

import { getHeadCellDriver, getDataCellDriver } from './index_driver'

describe('HeadCell', function() {
  let driver = null

  beforeEach(function() {
    driver = getHeadCellDriver()
    driver.when.created({
      content: 'Head cell',
    })
  })

  it('should match snapshot', function() {
    expect(driver.component).toMatchSnapshot()
  })

  it('should be a `th` element', function() {
    expect(driver.get.cellType()).toBe('th')
  })
})

describe('DataCell', function() {
  let driver = null

  beforeEach(function() {
    driver = getDataCellDriver()
    driver.when.created({
      children: 'Data entry',
    })
  })

  it('should match snapshot', function() {
    expect(driver.component).toMatchSnapshot()
  })

  it('should be a `th` element', function() {
    expect(driver.get.cellType()).toBe('td')
  })
})
