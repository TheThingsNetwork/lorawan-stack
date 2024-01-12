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

import React, { useCallback, useEffect, useState } from 'react'
import classnames from 'classnames'

import PropTypes from '@ttn-lw/lib/prop-types'

import MenuLink from './link'
import CollapsibleItem from './collapsible'

import style from './item.styl'

const handleItemClick = event => {
  if (event && event.target) {
    event.target.blur()
  }
}

const SideNavigationItem = props => {
  const { className, children, title, depth, icon, path, exact, isActive, isMinimized } = props
  const [isExpanded, setIsExpanded] = useState(false)

  const handleExpandCollapsableItem = useCallback(() => {
    setIsExpanded(isExpanded => !isExpanded)
    document.activeElement.blur()
  }, [])

  useEffect(() => {
    // Make sure that the item corresponding to the currently open path is expanded
    // on initial render, if applicable
    if (Boolean(children)) {
      const paths = React.Children.toArray(children).reduce(
        (paths, child) => [...paths, ...(Boolean(child) ? child.props.path : [])],
        [],
      )
      for (const path of paths) {
        if (location.pathname.includes(path)) {
          setIsExpanded(true)
          return
        }
      }
    }
  }, [children])

  return (
    <li className={classnames(className, style.item)}>
      {Boolean(children) ? (
        <CollapsibleItem
          title={title}
          icon={icon}
          onClick={handleExpandCollapsableItem}
          depth={depth}
          isActive={isActive}
          isExpanded={isExpanded}
          isMinimized={isMinimized}
          children={children}
          currentPathName={location.pathname}
          onDropdownItemsClick={handleItemClick}
        />
      ) : (
        <LinkItem
          title={title}
          icon={icon}
          exact={exact}
          path={path}
          depth={depth}
          onDropdownItemsClick={handleItemClick}
        />
      )}
    </li>
  )
}

SideNavigationItem.propTypes = {
  children: PropTypes.node,
  className: PropTypes.string,
  depth: PropTypes.number,
  /** A flag specifying whether the path of the linkable item should be matched exactly or not. */
  exact: PropTypes.bool,
  /** The name of the icon for the side navigation item. */
  icon: PropTypes.string,
  /** A flag specifying whether the side navigation item is active or not. */
  isActive: PropTypes.bool,
  isMinimized: PropTypes.bool,
  /** The path of the linkable side navigation item. */
  path: PropTypes.string,
  /** The title of the side navigation item. */
  title: PropTypes.message.isRequired,
}

SideNavigationItem.defaultProps = {
  className: undefined,
  children: undefined,
  exact: false,
  icon: undefined,
  isActive: false,
  depth: 0,
  path: undefined,
  isMinimized: false,
}

const LinkItem = ({ onClick, title, icon, exact, path }) => {
  const handleLinkItemClick = useCallback(
    event => {
      document.activeElement.blur()
      onClick(event)
    },
    [onClick],
  )

  return (
    <MenuLink path={path} title={title} icon={icon} onClick={handleLinkItemClick} exact={exact} />
  )
}

LinkItem.propTypes = {
  exact: PropTypes.bool.isRequired,
  icon: PropTypes.string,
  onClick: PropTypes.func,
  path: PropTypes.string,
  title: PropTypes.message.isRequired,
}

LinkItem.defaultProps = {
  icon: undefined,
  path: undefined,
  onClick: () => null,
}

export default SideNavigationItem
