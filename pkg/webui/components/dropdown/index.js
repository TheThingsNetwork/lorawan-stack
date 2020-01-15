// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import Icon from '../icon'
import Message from '../../lib/components/message'
import PropTypes from '../../lib/prop-types'

import style from './dropdown.styl'

const Dropdown = ({ className, children, larger, onItemsClick }) => (
  <ul
    onClick={onItemsClick}
    className={classnames(style.dropdown, className, { [style.larger]: larger })}
  >
    {children}
  </ul>
)

Dropdown.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  larger: PropTypes.bool,
  onItemsClick: PropTypes.func,
}

Dropdown.defaultProps = {
  className: undefined,
  larger: false,
  onItemsClick: () => null,
}

const DropdownItem = function({ icon, title, path, action, exact }) {
  const iconElement = icon && <Icon className={style.icon} icon={icon} nudgeUp />
  const ItemElement = action ? (
    <button onClick={action} onKeyPress={action} role="tab" tabIndex="0">
      {iconElement}
      <Message content={title} />
    </button>
  ) : (
    <NavLink activeClassName={style.active} to={path} exact={exact}>
      {iconElement}
      <Message content={title} />
    </NavLink>
  )
  return (
    <li className={style.dropdownItem} key={title.id || title}>
      {ItemElement}
    </li>
  )
}

DropdownItem.propTypes = {
  action: PropTypes.func,
  exact: PropTypes.bool,
  icon: PropTypes.string.isRequired,
  path: PropTypes.string,
  title: PropTypes.message.isRequired,
}

DropdownItem.defaultProps = {
  action: undefined,
  exact: false,
  path: undefined,
}

Dropdown.Item = DropdownItem

export default Dropdown
