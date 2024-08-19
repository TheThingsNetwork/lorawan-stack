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
import classnames from 'classnames'

import Icon from '@ttn-lw/components/icon'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './shortcut-item.styl'

const ShortcutItem = ({ icon, title, link, action, className }) =>
  action ? (
    <button onClick={action} className={classnames(style.shortcut, className)}>
      <div className="pos-relative w-full h-full">
        <div className={style.shortcutTitleWrapper}>
          <Icon icon={icon} className={style.icon} size={28} />
          <Message content={title} className={style.title} component="span" />
        </div>
      </div>
    </button>
  ) : (
    <Link to={link} className={classnames(style.shortcut, className)}>
      <div className="pos-relative w-full h-full">
        <div className={style.shortcutTitleWrapper}>
          <Icon icon={icon} className={style.icon} size={28} />
          <Message content={title} className={style.title} component="span" />
        </div>
      </div>
    </Link>
  )

ShortcutItem.propTypes = {
  action: PropTypes.func,
  className: PropTypes.string,
  icon: PropTypes.icon.isRequired,
  link: PropTypes.string.isRequired,
  title: PropTypes.message.isRequired,
}

ShortcutItem.defaultProps = {
  className: undefined,
  action: undefined,
}

export default ShortcutItem
