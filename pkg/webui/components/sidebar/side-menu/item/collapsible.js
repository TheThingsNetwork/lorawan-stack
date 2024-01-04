// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import Dropdown from '@ttn-lw/components/dropdown'
import Button from '@ttn-lw/components/button'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import SideNavigationList from '../list'

import style from './item.styl'

const CollapsibleItem = ({
  children,
  onClick,
  isExpanded,
  isMinimized,
  title,
  icon,
  depth,
  onDropdownItemsClick,
  currentPathName,
}) => {
  const subItems = children
    .filter(item => Boolean(item) && 'props' in item)
    .map(item => ({
      title: item.props.title,
      path: item.props.path,
      icon: item.props.icon,
    }))

  const subItemActive = subItems.some(item => currentPathName.includes(item.path))

  return (
    <>
      <Button
        className={classnames(style.link, {
          'j-center': isMinimized,
          'pl-cs-xs': !isMinimized,
          'pl-0': !isMinimized,
          [style.buttonActive]: isMinimized && subItemActive,
        })}
        onClick={onClick}
      >
        {icon && <Icon icon={icon} className={style.icon} />}
        {!isMinimized && (
          <>
            <Message content={title} className={style.message} />
            <Icon
              icon="keyboard_arrow_down"
              className={classnames(style.expandIcon, {
                [style.expandIconOpen]: isExpanded,
              })}
            />
          </>
        )}
        {isMinimized && (
          <div className={style.flyOutListContainer}>
            <Dropdown open className={style.flyOutList} onItemsClick={onDropdownItemsClick}>
              <Dropdown.HeaderItem title={title.defaultMessage} />
              {subItems.map(item => (
                <Dropdown.Item
                  key={item.path}
                  title={item.title}
                  path={item.path}
                  icon={item.icon}
                />
              ))}
            </Dropdown>
          </div>
        )}
      </Button>
      {!isMinimized && (
        <SideNavigationList depth={depth + 1} isExpanded={isExpanded}>
          {children}
        </SideNavigationList>
      )}
    </>
  )
}

CollapsibleItem.propTypes = {
  children: PropTypes.node,
  currentPathName: PropTypes.string.isRequired,
  depth: PropTypes.number.isRequired,
  icon: PropTypes.string,
  isExpanded: PropTypes.bool.isRequired,
  isMinimized: PropTypes.bool.isRequired,
  onClick: PropTypes.func.isRequired,
  onDropdownItemsClick: PropTypes.func,
  title: PropTypes.message.isRequired,
}

CollapsibleItem.defaultProps = {
  children: undefined,
  icon: undefined,
  onDropdownItemsClick: () => null,
}

export default CollapsibleItem
