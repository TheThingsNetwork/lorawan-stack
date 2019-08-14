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
import PropTypes from 'prop-types'
import classnames from 'classnames'
import bind from 'autobind-decorator'
import { NavLink } from 'react-router-dom'

import style from './tab.styl'

@bind
class Tab extends React.PureComponent {
  handleClick() {
    const { onClick, name, disabled } = this.props

    if (!disabled) {
      onClick(name)
    }
  }

  render() {
    const {
      className,
      onClick,
      name,
      active = false,
      disabled = false,
      narrow,
      children,
      link,
      exact = true,
      ...rest
    } = this.props

    const tabItemClassNames = classnames(className, style.tabItem, {
      [style.tabItemNarrow]: narrow,
      [style.tabItemActive]: !disabled && active,
      [style.tabItemDefault]: !disabled && !active,
      [style.tabItemDisabled]: disabled,
    })

    const Component = link ? NavLink : 'span'
    const props = {
      role: 'button',
      className: tabItemClassNames,
      children,
    }
    if (link) {
      props.exact = exact
      props.to = link
      props.activeClassName = style.tabItemActive
    } else {
      props.onClick = this.handleClick
    }

    return (
      <li {...rest} className={style.tab}>
        <Component {...props} children={children} />
      </li>
    )
  }
}

Tab.propTypes = {
  /**
   * A click handler to be called when the selected tab changes. Passes
   * the name of the new active tab as an argument.
   */
  onClick: PropTypes.func,
  /** A flag specifying whether the tab is active */
  active: PropTypes.bool,
  /** A flag specifying whether the tab is disabled */
  disabled: PropTypes.bool,
  /** The name of the tab */
  name: PropTypes.string.isRequired,
}

export default Tab
