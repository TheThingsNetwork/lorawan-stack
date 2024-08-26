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

import React, { useCallback, useState } from 'react'
import classnames from 'classnames'

import SafeInspector from '@ttn-lw/components/safe-inspector'
import Tooltip from '@ttn-lw/components/tooltip'
import Icon, { IconInfoCircle, IconChevronDown, IconChevronUp } from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './data-sheet.styl'

const DataSheet = ({ className, data }) => (
  <div className={classnames(style.dataSheet, className)}>
    {data.map((group, index) => (
      <DataSheetSection
        key={`${group.header}_${index}`}
        dataLength={data.length}
        group={group}
        index={index}
      />
    ))}
  </div>
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

const DataSheetSection = ({ dataLength, group, index }) => {
  const hasHeader = Boolean(group.header)
  const [isOpen, setIsOpen] = useState(!hasHeader)

  const toggleOpen = useCallback(() => {
    setIsOpen(prevOpen => !prevOpen)
  }, [])

  return (
    <div>
      {Boolean(group.header) && (
        <div className={style.dataSheetHeader} onClick={toggleOpen}>
          <Message content={group.header} />
          <Icon icon={isOpen ? IconChevronUp : IconChevronDown} className={style.icon} />
        </div>
      )}
      <div
        className={classnames({
          [style.dataSheetSectionCollapsed]: !isOpen,
        })}
      >
        {group.items.length > 0 ? (
          group.items.map((item, itemIndex) => {
            if (!item) {
              return null
            }
            const keyId = typeof item.key === 'object' ? item.key.id : item.key
            const subItems = item.subItems
              ? item.subItems.map((subItem, subIndex) => (
                  <div
                    key={`${keyId}_${index}`}
                    className={classnames(style.dataSheetRowContent, 'pl-cs-l')}
                  >
                    <DataSheetRow item={subItem} key={`${keyId}_${index}_${subIndex}`} />
                  </div>
                ))
              : null

            return (
              <div
                key={`${keyId}_${index}`}
                className={classnames(style.dataSheetRowContent, {
                  [style.lastGroupLastItem]:
                    index === dataLength - 1 && itemIndex === group.items.length - 1,
                })}
              >
                <DataSheetRow item={item} />
                {subItems}
              </div>
            )
          })
        ) : (
          <div className={style.dataSheetRowContent}>
            <Message content={group.emptyMessage || sharedMessages.noData} />
          </div>
        )}
      </div>
      {index !== dataLength - 1 && <div className={style.dataSheetDivider} />}
    </div>
  )
}

DataSheetSection.propTypes = {
  dataLength: PropTypes.number.isRequired,
  group: PropTypes.shape({
    emptyMessage: PropTypes.message,
    header: PropTypes.message.isRequired,
    items: PropTypes.arrayOf(
      PropTypes.shape({
        enableUint32: PropTypes.bool,
        key: PropTypes.message,
        value: PropTypes.message,
        type: PropTypes.string,
        sensitive: PropTypes.bool,
        subItems: PropTypes.arrayOf(PropTypes.shape({})),
      }),
    ),
  }).isRequired,
  index: PropTypes.number.isRequired,
}

const DataSheetRow = ({ item }) => {
  const isSafeInspector = item.type === 'byte' || item.type === 'code'

  return (
    <>
      {item.key && (
        <div className={style.dataSheetRowHeading}>
          <Message content={item.key} />
          {item.tooltipMessage && (
            <Tooltip content={<Message content={item.tooltipMessage} />} small>
              <Icon icon={IconInfoCircle} className={style.tooltipIcon} size={16} />
            </Tooltip>
          )}
        </div>
      )}
      <div className={style.dataSheetRowContentValue}>
        {item.value && isSafeInspector ? (
          <SafeInspector
            hideable={false || item.sensitive}
            isBytes={item.type === 'byte'}
            data={item.value}
            enableUint32={item.enableUint32}
            small
          />
        ) : (
          item.value || (
            <Message className={style.notAvailable} content={sharedMessages.notAvailable} />
          )
        )}
      </div>
    </>
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
}

export default DataSheet
