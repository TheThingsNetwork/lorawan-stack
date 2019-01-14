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
import classnames from 'classnames'

import SafeInspector from '../safe-inspector'

import style from './data-sheet.styl'

const DataSheet = function ({ data }) {

  return (
    <table className={style.table}>
      <tbody>
        { data.map(function (group, index) {
          return (
            <React.Fragment key={`${group.header}_${index}`}>
              <tr className={style.groupHeading}><th>{group.header}</th><td /></tr>
              { group.items.map( function (item) {
                const isSafeInspector = item.type === 'byte' || item.type === 'code'
                return (
                  <tr className={classnames({ [style.taller]: item.safe })} key={`${item.key}_${index}`}>
                    <th>{item.key}</th>
                    <td>{isSafeInspector ? (
                      <SafeInspector
                        hideable={false || item.sensitive}

                        isBytes={item.type === 'byte'}
                        small
                        data={item.value}
                      />
                    )
                      : item.value}</td>
                  </tr>
                )
              }
              )}
            </React.Fragment>
          )
        })}
      </tbody>
    </table>
  )
}

DataSheet.propTypes = {}

export default DataSheet
