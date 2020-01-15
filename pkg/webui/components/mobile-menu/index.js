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
import { defineMessages } from 'react-intl'
import classnames from 'classnames'

import Icon from '../icon'
import Button from '../button'
import Message from '../../lib/components/message'
import Dropdown from '../dropdown'
import sharedMessages from '../../lib/shared-messages'
import PropTypes from '../../lib/prop-types'

import style from './mobile-navigation-dropdown.styl'

const m = defineMessages({
  loggedInAs: 'Logged in as <b>{userId}</b>',
})

const MobileMenu = ({ className, children, user, onItemsClick, onLogout }) => {
  return (
    <div className={classnames(className, style.container)}>
      <Dropdown
        className={style.mobileDropdown}
        itemClassName={style.mobileDropdownItem}
        onItemsClick={onItemsClick}
        larger
      >
        {children}
      </Dropdown>
      <div className={style.userHeader}>
        <div className={style.userMessage}>
          <Icon className={style.userIcon} icon="person" nudgeUp />
          <Message
            className={style.userMessage}
            content={m.loggedInAs}
            values={{ userId: user.ids.user_id, b: (...chunks) => <b key="1"> {chunks}</b> }}
          />
        </div>
        <div>
          <Button message={sharedMessages.logout} icon="logout" onClick={onLogout} naked />
        </div>
      </div>
    </div>
  )
}

MobileMenu.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  onItemsClick: PropTypes.func.isRequired,
  onLogout: PropTypes.func.isRequired,
  user: PropTypes.user.isRequired,
}

MobileMenu.defaultProps = {
  className: undefined,
}

export default MobileMenu
