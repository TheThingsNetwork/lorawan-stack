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

import React, { Fragment } from 'react'
import classnames from 'classnames'
import bind from 'autobind-decorator'
import PropTypes from '../../../../lib/prop-types'

import SideNavigationList from '../list'
import NavigationLink from '../../link'
import Message from '../../../../lib/components/message'
import Icon from '../../../icon'

import style from './item.styl'

@bind
class SideNavigationItem extends React.PureComponent {
  onExpandCollapsableItem() {
    this.props.onExpand(false)
  }

  onExpandLinkItem() {
    this.props.onExpand(true)
  }

  render() {
    const {
      className,
      title,
      depth,
      icon = null,
      path = null,
      exact = true,
      onExpand,
      isCollapsable = false,
      isMinimized,
      isExpanded,
      isActive,
      items,
    } = this.props

    return (
      <li
        className={classnames(className, style.item, {
          [style.itemMinimized]: isMinimized,
        })}
      >
        {isCollapsable ? (
          <CollapsableItem
            title={title}
            icon={icon}
            onExpand={onExpand}
            onClick={this.onExpandCollapsableItem}
            depth={depth}
            items={items}
            isActive={isActive}
            isExpanded={isExpanded}
            isMinimized={isMinimized}
          />
        ) : (
          <LinkItem
            title={title}
            icon={icon}
            exact={exact}
            path={path}
            depth={depth}
            onExpand={this.onExpandLinkItem}
          />
        )}
      </li>
    )
  }
}

const CollapsableItem = ({
  onClick,
  onExpand,
  isActive,
  isExpanded,
  isMinimized,
  title,
  icon,
  depth,
  items,
}) => (
  <Fragment>
    <button
      className={classnames(style.button, {
        [style.buttonActive]: isActive,
      })}
      type="button"
      data-hook="side-nav-item-button"
      onClick={onClick}
    >
      {icon && <Icon icon={icon} className={style.icon} />}
      <Message content={title} className={style.message} />
      <Icon
        icon="keyboard_arrow_down"
        className={classnames(style.expandIcon, {
          [style.expandIconOpen]: isExpanded,
        })}
      />
    </button>
    <SideNavigationList
      isMinimized={isMinimized}
      depth={depth + 1}
      items={items}
      isExpanded={isExpanded}
      onItemExpand={onExpand}
    />
  </Fragment>
)

const LinkItem = ({ title, icon, exact, path, onExpand }) => (
  <NavigationLink
    className={style.link}
    activeClassName={style.linkActive}
    exact={exact}
    path={path}
    onClick={onExpand}
    data-hook="side-nav-item-link"
  >
    {icon && <Icon icon={icon} className={style.icon} />}
    <Message content={title} className={style.message} />
  </NavigationLink>
)

SideNavigationItem.propTypes = {
  /** The title of the side navigation item */
  title: PropTypes.message.isRequired,
  /** The name of the icon for the side navigation item */
  icon: PropTypes.string,
  /** The path of the linkable side navigation item */
  path: PropTypes.string,
  /** A flag specifying whether the path of the linkable item should be matched exactly or not */
  exact: PropTypes.bool,
  /** A flag specifying whether the side navigation item is active or not */
  isActive: PropTypes.bool.isRequired,
  /** A flag specifying whether the side navigation item is minimized or not */
  isMinimized: PropTypes.bool.isRequired,
  /**
   * A flag specifying whether the side navigation item is composed of multiple
   * entries and is collapsable/expandable
   */
  isCollapsable: PropTypes.bool.isRequired,
  /** A flag specifying whether the side navigation item is expanded */
  isExpanded: PropTypes.bool.isRequired,
  /** Function to be called when the item gets selected */
  onExpand: PropTypes.func.isRequired,
}

export default SideNavigationItem
