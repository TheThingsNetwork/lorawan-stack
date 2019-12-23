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

import { NavLink } from 'react-router-dom'

import Link from '../../link'

import PropTypes from '../../../lib/prop-types'
import style from './link.styl'

const NavigationLink = function({
  className,
  children,
  path,
  exact,
  activeClassName,
  onClick,
  ...rest
}) {
  return (
    <NavLink
      to={path}
      exact={exact}
      className={classnames(className, style.link)}
      activeClassName={activeClassName}
      onClick={onClick}
      {...rest}
    >
      {children}
    </NavLink>
  )
}

const NavigationAnchorLink = function({ className, children, path, ...rest }) {
  return (
    <Link.BaseAnchor
      href={path}
      className={classnames(className, style.link)}
      children={children}
    />
  )
}

NavigationLink.propTypes = {
  activeClassName: PropTypes.string,
  children: PropTypes.node,
  className: PropTypes.string,
  /**
   * Boolean flag identifying whether the path should
   * be matched exactly
   */
  exact: PropTypes.bool.isRequired,
  onClick: PropTypes.func,
  /** The path for a link */
  path: PropTypes.string.isRequired,
  /** The name of a css class to be applied on the active tab */
}

NavigationLink.defaultProps = {
  activeClassName: undefined,
  children: undefined,
  className: undefined,
  onClick: () => null,
}

NavigationAnchorLink.propTypes = {
  children: PropTypes.node,
  className: PropTypes.string,
  path: PropTypes.string.isRequired,
}

NavigationAnchorLink.defaultProps = {
  className: undefined,
  children: undefined,
}
export { NavigationLink as default, NavigationAnchorLink }
