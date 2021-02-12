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
import { withRouter } from 'react-router-dom'

import Dropdown from '@ttn-lw/components/dropdown'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import SideNavigationList from '../list'
import NavigationLink from '../../link'
import SideNavigationContext from '../context'

import style from './item.styl'

export class SideNavigationItem extends React.PureComponent {
  static contextType = SideNavigationContext

  static propTypes = {
    children: PropTypes.node,
    className: PropTypes.string,
    depth: PropTypes.number,
    /** A flag specifying whether the path of the linkable item should be matched exactly or not. */
    exact: PropTypes.bool,
    /** The name of the icon for the side navigation item. */
    icon: PropTypes.string,
    /** A flag specifying whether the side navigation item is active or not. */
    isActive: PropTypes.bool,
    location: PropTypes.location.isRequired,
    /** The path of the linkable side navigation item. */
    path: PropTypes.string,
    /** The title of the side navigation item. */
    title: PropTypes.message.isRequired,
  }

  static defaultProps = {
    className: undefined,
    children: undefined,
    exact: false,
    icon: undefined,
    isActive: false,
    depth: 0,
    path: undefined,
  }

  state = {
    isExpanded: false,
  }

  handleExpandCollapsableItem() {
    this.setState({ isExpanded: !this.state.isExpanded })
    document.activeElement.blur()
  }

  componentDidMount() {
    // Make sure that the item corresponding to the currently open path is expanded
    // on initial render, if applicable
    const { location, children } = this.props
    if (Boolean(children)) {
      const paths = React.Children.toArray(children).reduce(
        (paths, child) => [...paths, ...(Boolean(child) ? child.props.path : [])],
        [],
      )
      for (const path of paths) {
        if (location.pathname.startsWith(path)) {
          this.setState({ isExpanded: true })
          return
        }
      }
    }
  }

  handleItemClick = event => {
    if (event && event.target) {
      event.target.blur()
    }
  }

  render() {
    const { className, children, title, depth, icon, path, exact, isActive, location } = this.props
    const { isExpanded } = this.state
    const { isMinimized, onLeafItemClick } = this.context

    return (
      <li
        className={classnames(className, style.item, {
          [style.itemMinimized]: isMinimized,
        })}
      >
        {Boolean(children) ? (
          <CollapsableItem
            title={title}
            icon={icon}
            onClick={this.handleExpandCollapsableItem}
            depth={depth}
            isActive={isActive}
            isExpanded={isExpanded}
            isMinimized={isMinimized}
            children={children}
            currentPathName={location.pathname}
            onDropdownItemsClick={this.handleItemClick}
          />
        ) : (
          <LinkItem
            onClick={onLeafItemClick}
            title={title}
            icon={icon}
            exact={exact}
            path={path}
            depth={depth}
            onDropdownItemsClick={this.handleItemClick}
          />
        )}
      </li>
    )
  }
}

const CollapsableItem = ({
  children,
  onClick,
  isActive,
  isExpanded,
  title,
  icon,
  depth,
  currentPathName,
  onDropdownItemsClick,
}) => {
  const subItems = children.map(item => ({
    title: item.props.title,
    path: item.props.path,
    icon: item.props.icon,
  }))

  const subItemActive = subItems.some(item => item.path === currentPathName)

  return (
    <>
      <button
        className={classnames(style.button, {
          [style.buttonActive]: isActive,
          [style.linkActive]: subItemActive,
        })}
        type="button"
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
      <Dropdown className={style.flyOutList} onItemsClick={onDropdownItemsClick}>
        <Dropdown.HeaderItem title={title.defaultMessage} />
        {subItems.map(item => (
          <Dropdown.Item key={item.path} title={item.title} path={item.path} icon={item.icon} />
        ))}
      </Dropdown>
      <SideNavigationList depth={depth + 1} isExpanded={isExpanded} className={style.subItems}>
        {children}
      </SideNavigationList>
    </>
  )
}

CollapsableItem.propTypes = {
  children: PropTypes.node,
  currentPathName: PropTypes.string.isRequired,
  depth: PropTypes.number.isRequired,
  icon: PropTypes.string,
  isActive: PropTypes.bool.isRequired,
  isExpanded: PropTypes.bool.isRequired,
  onClick: PropTypes.func.isRequired,
  onDropdownItemsClick: PropTypes.func,
  title: PropTypes.message.isRequired,
}

CollapsableItem.defaultProps = {
  children: undefined,
  icon: undefined,
  onDropdownItemsClick: () => null,
}

const LinkItem = ({ onClick, title, icon, exact, path, onDropdownItemsClick }) => {
  const handleLinkItemClick = React.useCallback(
    event => {
      document.activeElement.blur()
      onClick(event)
    },
    [onClick],
  )

  return (
    <>
      <NavigationLink
        onClick={handleLinkItemClick}
        className={style.link}
        activeClassName={style.linkActive}
        exact={exact}
        path={path}
      >
        {icon && <Icon icon={icon} className={style.icon} />}
        <Message content={title} className={style.message} />
      </NavigationLink>
      <Dropdown className={style.flyOutList} onItemsClick={onDropdownItemsClick}>
        <Dropdown.Item title={title} path={path} showActive={false} icon={''} tabIndex="-1" />
      </Dropdown>
    </>
  )
}

LinkItem.propTypes = {
  exact: PropTypes.bool.isRequired,
  icon: PropTypes.string,
  onClick: PropTypes.func,
  onDropdownItemsClick: PropTypes.func,
  path: PropTypes.string,
  title: PropTypes.message.isRequired,
}

LinkItem.defaultProps = {
  icon: undefined,
  path: undefined,
  onClick: () => null,
  onDropdownItemsClick: () => null,
}

export default withRouter(bind(SideNavigationItem))
