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
import { withRouter } from 'react-router-dom'

import PropTypes from '../../../../lib/prop-types'

import SideNavigationList from '../list'
import NavigationLink from '../../link'
import Message from '../../../../lib/components/message'
import Icon from '../../../icon'
import SideNavigationContext from '../context'

import style from './item.styl'

export class SideNavigationItem extends React.PureComponent {
  static contextType = SideNavigationContext

  static propTypes = {
    children: PropTypes.node,
    className: PropTypes.string,
    depth: PropTypes.number,
    /** A flag specifying whether the path of the linkable item should be matched exactly or not */
    exact: PropTypes.bool,
    /** The name of the icon for the side navigation item */
    icon: PropTypes.string,
    /** A flag specifying whether the side navigation item is active or not */
    isActive: PropTypes.bool,
    location: PropTypes.location.isRequired,
    /** The path of the linkable side navigation item */
    path: PropTypes.string,
    /** The title of the side navigation item */
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
  }

  componentDidMount() {
    // Make sure that the item corresponding to the currently open path is expanded
    // on initial render, if applicable
    const { location, children } = this.props
    if (Boolean(children)) {
      const paths = React.Children.map(children, child => child.props.path)
      for (const path of paths) {
        if (location.pathname.startsWith(path)) {
          this.setState({ isExpanded: true })
          return
        }
      }
    }
  }

  render() {
    const { className, children, title, depth, icon, path, exact, isActive } = this.props
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
          />
        ) : (
          <LinkItem
            onClick={onLeafItemClick}
            title={title}
            icon={icon}
            exact={exact}
            path={path}
            depth={depth}
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
  isMinimized,
  title,
  icon,
  depth,
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
    <SideNavigationList isMinimized={isMinimized} depth={depth + 1} isExpanded={isExpanded}>
      {children}
    </SideNavigationList>
  </Fragment>
)

CollapsableItem.propTypes = {
  children: PropTypes.node,
  depth: PropTypes.number.isRequired,
  icon: PropTypes.string,
  isActive: PropTypes.bool.isRequired,
  isExpanded: PropTypes.bool.isRequired,
  isMinimized: PropTypes.bool,
  onClick: PropTypes.func.isRequired,
  title: PropTypes.message.isRequired,
}

CollapsableItem.defaultProps = {
  children: undefined,
  icon: undefined,
  isMinimized: false,
}

const LinkItem = ({ onClick, title, icon, exact, path }) => (
  <NavigationLink
    onClick={onClick}
    className={style.link}
    activeClassName={style.linkActive}
    exact={exact}
    path={path}
  >
    {icon && <Icon icon={icon} className={style.icon} />}
    <Message content={title} className={style.message} />
  </NavigationLink>
)

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
}

export default withRouter(bind(SideNavigationItem))
