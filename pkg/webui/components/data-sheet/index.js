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

import SafeInspector from '@ttn-lw/components/safe-inspector'
import Tooltip from '@ttn-lw/components/tooltip'
import Icon, { IconInfoCircle } from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './data-sheet.styl'

const DataSheet = ({ className, data }) => (
  <table className={classnames(className, style.table)}>
    <tbody>
      {data.map((group, index) => (
        <React.Fragment key={`${group.header}_${index}`}>
          <tr className={style.groupHeading}>
            <th>
              <Message content={group.header} />
            </th>
          </tr>
          {group.items.length > 0 ? (
            group.items.map(item => {
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
                <Message content={group.emptyMessage || sharedMessages.noData} />
              </th>
            </tr>
          )}
        </React.Fragment>
      ))}
    </tbody>
  </table>
)

DataSheet.propTypes = {
  className: PropTypes.string,
  /** A list of entries for the sheet. */
  data: PropTypes.arrayOf(
    PropTypes.shape({
      emptyMessage: PropTypes.message,
      /** The title of the item group. */
      header: PropTypes.message.isRequired,
      /** A list of items for the group. */
      items: PropTypes.arrayOf(
        PropTypes.shape({
          /** Whether uint32_t notation should be enabled for byte representation. */
          enableUint32: PropTypes.bool,
          /** The key of the item. */
          key: PropTypes.message,
          /** The value of the item. */
          value: PropTypes.message,
          /** The type of the item, 'code', 'byte' or 'text' (default). */
          type: PropTypes.string,
          /** Whether this 'code' or 'byte' item should be hidden by default. */
          sensitive: PropTypes.bool,
          /** Optional subitems of this item (same shape as item, but no deeper hierarchies). */
          subItems: PropTypes.arrayOf(PropTypes.shape({})),
        }),
      ),
    }),
  ).isRequired,
}

DataSheet.defaultProps = {
  className: undefined,
}

const DataSheetRow = ({ item, sub }) => {
  const isSafeInspector = item.type === 'byte' || item.type === 'code'
  const rowStyle = classnames({
    [style.sub]: sub,
  })

  return (
    <tr className={rowStyle}>
      {item.key && (
        <th>
          <div className="d-flex al-center gap-cs-xxs">
            <Message content={item.key} />
            {item.tooltipMessage && (
              <Tooltip content={<Message content={item.tooltipMessage} />}>
                <Icon icon={IconInfoCircle} className="c-text-neutral-semilight" size={17} />
              </Tooltip>
            )}
          </div>
        </th>
      )}
      <td>
        {item.value && isSafeInspector ? (
          <SafeInspector
            hideable={false || item.sensitive}
            isBytes={item.type === 'byte'}
            data={item.value}
            enableUint32={item.enableUint32}
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
    /** Whether uint32_t notation should be enabled for byte representation. */
    enableUint32: PropTypes.bool,
    /** The key of the item. */
    key: PropTypes.message,
    /** The value of the item. */
    value: PropTypes.message,
    /** The type of the item, 'code', 'byte' or 'text' (default). */
    type: PropTypes.string,
    tooltipMessage: PropTypes.message,
    /** Whether this 'code' or 'byte' item should be hidden by default. */
    sensitive: PropTypes.bool,
    /** Optional subitems of this item (same shape as item, but no deeper hierarchies). */
    subItems: PropTypes.arrayOf(PropTypes.shape({})),
  }).isRequired,
  sub: PropTypes.bool,
}

DataSheetRow.defaultProps = {
  sub: false,
}

export default DataSheet
