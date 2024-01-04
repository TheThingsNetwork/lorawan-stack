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

import React from 'react'
import { NavLink } from 'react-router-dom'
import classNames from 'classnames'

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './shortcut-item.styl'

const ShortcutItem = ({ icon, title, link }) => (
  <NavLink
    to={link}
    className={classNames(
      style.shortcut,
      'd-flex gap-cs-m direction-column al-center j-center p-vert-cs-s p-sides-cs-m',
    )}
  >
    <Icon icon="add" className={style.addIcon} />
    <Icon icon={icon} className={style.icon} />
    <Message content={title} className={style.title} />
  </NavLink>
)

ShortcutItem.propTypes = {
  icon: PropTypes.string.isRequired,
  link: PropTypes.string.isRequired,
  title: PropTypes.message.isRequired,
}

export default ShortcutItem
