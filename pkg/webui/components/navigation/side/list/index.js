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
import bind from 'autobind-decorator'
import PropTypes from '../../../../lib/prop-types'

import SideNavigationItem from '../item'

import style from './list.styl'

@bind
class SideNavigationList extends React.Component {
  static propTypes = {
    className: PropTypes.string,
    /** The depth of the current list starting at 0 for the root list */
    depth: PropTypes.number,
    /**
     * A flag specifying whether the side navigation list is expanded or not.
     * Applicable to nested lists.
     */
    isExpanded: PropTypes.bool,
    /** A flag specifying whether the side navigation list of items is minimized or not */
    isMinimized: PropTypes.bool.isRequired,
    /** A list of items to be displayed within the side navigation list */
    items: PropTypes.arrayOf(
      PropTypes.oneOfType([
        PropTypes.link,
        PropTypes.shape({
          title: PropTypes.message.isRequired,
          icon: PropTypes.string,
          nested: PropTypes.bool.isRequired,
          items: PropTypes.arrayOf(PropTypes.link).isRequired,
          hidden: PropTypes.bool,
        }),
      ]),
    ).isRequired,
    itemsExpanded: PropTypes.shape({}),
    /** Function to be called when an side navigation item gets selected */
    onItemExpand: PropTypes.func,
    /**
     * A map of expanded items, where:
     *  - The key: index of the item
     *  - The value: an object consisting of:
     *    - isOpen - boolean flag specifying whether the item is opened or not
     *    - isLink - boolean flag specifying whether a link is selected within
     *                this opened item
     */
  }

  static defaultProps = {
    className: undefined,
    depth: 0,
    itemsExpanded: {},
    isExpanded: false,
    onItemExpand: () => null,
  }

  onRootExpand(index) {
    const { onItemExpand } = this.props

    return function(isLink) {
      onItemExpand(index, isLink)
    }
  }

  render() {
    const {
      className,
      items,
      isMinimized,
      onItemExpand,
      isExpanded,
      depth,
      itemsExpanded = {},
    } = this.props

    const onRootExpand = this.onRootExpand
    const isRoot = depth === 0
    const listClassNames = classnames(className, style.list, {
      [style.listNested]: !isRoot,
      [style.listExpanded]: isExpanded,
    })
    return (
      <ul className={listClassNames}>
        {items.map(function(item, index) {
          const itemState = itemsExpanded[index] || {}
          const { title, icon, path, exact = true, nested = false, items = [], hidden } = item

          if (hidden) return null

          const { isOpen = false, isLink = false } = itemState

          const isActive = nested && isLink
          const isExpanded = !isMinimized && isOpen
          const onExpand = isRoot ? onRootExpand(index) : onItemExpand

          return (
            <SideNavigationItem
              key={index}
              title={title}
              icon={icon}
              path={path}
              exact={exact}
              depth={depth}
              onExpand={onExpand}
              isMinimized={isMinimized}
              isCollapsable={nested}
              isExpanded={isExpanded}
              isActive={isActive}
              items={items}
            />
          )
        })}
      </ul>
    )
  }
}

export default SideNavigationList
