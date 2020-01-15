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

import Icon from '../icon'
import Message from '../../lib/components/message'
import Link from '../link'
import PropTypes from '../../lib/prop-types'

import style from './dropdown.styl'

const Dropdown = ({ className, children }) => (
  <ul className={classnames(style.dropdown, className)}>{children}</ul>
)

Dropdown.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
}

Dropdown.defaultProps = {
  className: undefined,
}

const DropdownItem = function({ icon, title, path, action }) {
  const iconElement = icon && <Icon className={style.icon} icon={icon} />
  const ItemElement = action ? (
    <button onClick={action} onKeyPress={action} role="tab" tabIndex="0">
      {iconElement}
      <Message content={title} />
    </button>
  ) : (
    <Link to={path}>
      {iconElement}
      <Message content={title} />
    </Link>
  )
  return (
    <li className={style.dropdownItem} key={title.id || title}>
      {ItemElement}
    </li>
  )
}

DropdownItem.propTypes = {
  action: PropTypes.func,
  icon: PropTypes.string.isRequired,
  path: PropTypes.string,
  title: PropTypes.message.isRequired,
}

DropdownItem.defaultProps = {
  action: undefined,
  path: undefined,
}

Dropdown.Item = DropdownItem

export default Dropdown
