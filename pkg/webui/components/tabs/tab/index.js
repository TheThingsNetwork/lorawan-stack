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

import React, { useCallback } from 'react'
import classnames from 'classnames'
import { NavLink } from 'react-router-dom'

import Tooltip from '@ttn-lw/components/tooltip'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './tab.styl'

const Tab = props => {
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
    tabClassName,
    toggleStyle,
    tooltip,
    ...rest
  } = props

  const handleClick = useCallback(() => {
    if (!disabled) {
      onClick(name)
    }
  }, [disabled, name, onClick])

  const tabClassNames = classnames(tabClassName, style.tab)

  const tabItemClassNames = classnames(className, style.tabItem, {
    [style.tabItemNarrow]: narrow,
    [style.tabItemActive]: !toggleStyle && !disabled && active,
    [style.tabItemDefault]: !toggleStyle && !disabled && !active,
    [style.tabItemDisabled]: disabled,
    [style.tabItemToggleStyle]: toggleStyle,
    [style.tabItemToggleStyleActive]: toggleStyle && !disabled && active,
  })

  // There is no support for disabled on anchors in html and hence in
  // `react-router`. So, do not render the link component if the tab is
  // disabled, but render regular tab item instead.
  const canRenderLink = link && !disabled

  const Component = canRenderLink ? NavLink : 'span'
  const componentProps = {
    role: 'button',
    className: tabItemClassNames,
    children,
  }

  if (canRenderLink) {
    componentProps.end = exact
    componentProps.to = link
    componentProps.className = ({ isActive }) =>
      classnames(tabItemClassNames, style.tabItem, {
        [style.tabItemActive]: !disabled && isActive,
      })
  } else {
    componentProps.onClick = handleClick
  }

  const wrappedChildren = tooltip ? (
    <Tooltip content={<Message content={tooltip} />} delay={0} small>
      <Component {...componentProps} children={children} />
    </Tooltip>
  ) : (
    <Component {...componentProps} children={children} />
  )

  return (
    <li {...rest} className={tabClassNames}>
      {wrappedChildren}
    </li>
  )
}

Tab.propTypes = {
  /** A flag specifying whether the tab is active. */
  active: PropTypes.bool,
  children: PropTypes.node,
  className: PropTypes.string,
  /** A flag specifying whether the tab is disabled. */
  disabled: PropTypes.bool,
  exact: PropTypes.bool,
  link: PropTypes.string,
  /** The name of the tab. */
  name: PropTypes.string.isRequired,
  narrow: PropTypes.bool,
  /**
   * A click handler to be called when the selected tab changes. Passes the
   * name of the new active tab as an argument.
   */
  onClick: PropTypes.func,
  tabClassName: PropTypes.string,
  /** A flag specifying whether the tab should render a toggle style. */
  toggleStyle: PropTypes.bool,
  /** A tooltip to be displayed on hover. */
  tooltip: PropTypes.message,
}

Tab.defaultProps = {
  children: undefined,
  className: undefined,
  tabClassName: undefined,
  link: undefined,
  onClick: () => null,
  active: false,
  disabled: false,
  narrow: false,
  exact: true,
  toggleStyle: false,
  tooltip: undefined,
}

export default Tab
