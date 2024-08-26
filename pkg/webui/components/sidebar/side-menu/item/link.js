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

import React, { useCallback, useRef } from 'react'
import { NavLink } from 'react-router-dom'
import classnames from 'classnames'

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './item.styl'

const MenuLink = ({ icon, title, path, onClick, exact, disabled, external }) => {
  const ref = useRef()

  const className = useCallback(
    ({ isActive }) =>
      classnames(style.link, style.container, {
        [style.active]: isActive,
        [style.disabled]: disabled,
      }),
    [disabled],
  )

  return (
    <NavLink
      to={path}
      className={className}
      end={exact}
      onClick={onClick}
      ref={ref}
      target={external ? '_blank' : undefined}
    >
      {icon && <Icon icon={icon} className={classnames(style.icon)} />}{' '}
      <Message className={style.title} content={title} />
    </NavLink>
  )
}

MenuLink.propTypes = {
  disabled: PropTypes.bool,
  exact: PropTypes.bool.isRequired,
  external: PropTypes.bool,
  icon: PropTypes.icon,
  onClick: PropTypes.func,
  path: PropTypes.string.isRequired,
  title: PropTypes.message.isRequired,
}

MenuLink.defaultProps = {
  icon: undefined,
  onClick: () => null,
  disabled: false,
  external: false,
}

export default MenuLink
