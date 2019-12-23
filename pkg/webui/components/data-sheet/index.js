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
import { defineMessages } from 'react-intl'

import SafeInspector from '../safe-inspector'
import Message from '../../lib/components/message'

import PropTypes from '../../lib/prop-types'
import sharedMessages from '../../lib/shared-messages'

import style from './data-sheet.styl'

const m = defineMessages({
  noData: 'No data available',
})

const DataSheet = function({ className, data }) {
  return (
    <table className={classnames(className, style.table)}>
      <tbody>
        {data.map(function(group, index) {
          return (
            <React.Fragment key={`${group.header}_${index}`}>
              <tr className={style.groupHeading}>
                <th>
                  <Message content={group.header} />
                </th>
              </tr>
              {group.items.length > 0 ? (
                group.items.map(function(item) {
                  if (!item) {
                    return null
                  }
                  const keyId = typeof item.key === 'object' ? item.key.id : item.key
                  const subItems = item.subItems
                    ? item.subItems.map((subItem, subIndex) => (
                        <DataSheetRow sub item={subItem} key={`${keyId}_${index}_${subIndex}`} />
                      ))
                    : null

                  return (
                    <React.Fragment key={`${keyId}_${index}`}>
                      <DataSheetRow item={item} />
                      {subItems}
                    </React.Fragment>
                  )
                })
              ) : (
                <tr>
                  <th colSpan={2}>
                    <Message content={m.noData} />
                  </th>
                </tr>
              )}
            </React.Fragment>
          )
        })}
      </tbody>
    </table>
  )
}

DataSheet.propTypes = {
  className: PropTypes.string,
  /** A list of entries for the sheet */
  data: PropTypes.arrayOf(
    PropTypes.shape({
      /** The title of the item group */
      header: PropTypes.message.isRequired,
      /** A list of items for the group */
      items: PropTypes.arrayOf(
        PropTypes.shape({
          /** The key of the item */
          key: PropTypes.message,
          /** The value of the item */
          value: PropTypes.message,
          /** The type of the item, 'code', 'byte' or 'text' (default) */
          type: PropTypes.string,
          /** Whether this 'code' or 'byte' item should be hidden by default */
          sensitive: PropTypes.bool,
          /** Optional subitems of this item (same shape as item, but no deeper
           * hierarchies) */
          subItems: PropTypes.arrayOf(PropTypes.object),
        }),
      ),
    }),
  ).isRequired,
}

DataSheet.defaultProps = {
  className: undefined,
}

const DataSheetRow = function({ item, sub }) {
  const isSafeInspector = item.type === 'byte' || item.type === 'code'
  const rowStyle = classnames({
    [style.sub]: sub,
  })

  return (
    <tr className={rowStyle}>
      <th>
        <Message content={item.key} />
      </th>
      <td>
        {item.value && isSafeInspector ? (
          <SafeInspector
            hideable={false || item.sensitive}
            isBytes={item.type === 'byte'}
            small
            data={item.value}
          />
        ) : (
          item.value || (
            <Message className={style.notAvailable} content={sharedMessages.notAvailable} />
          )
        )}
      </td>
    </tr>
  )
}

DataSheetRow.propTypes = {
  item: PropTypes.shape({
    /** The key of the item */
    key: PropTypes.message,
    /** The value of the item */
    value: PropTypes.message,
    /** The type of the item, 'code', 'byte' or 'text' (default) */
    type: PropTypes.string,
    /** Whether this 'code' or 'byte' item should be hidden by default */
    sensitive: PropTypes.bool,
    /** Optional subitems of this item (same shape as item, but no deeper
     * hierarchies) */
    subItems: PropTypes.arrayOf(PropTypes.object),
  }).isRequired,
  sub: PropTypes.bool,
}

DataSheetRow.defaultProps = {
  sub: false,
}

export default DataSheet
