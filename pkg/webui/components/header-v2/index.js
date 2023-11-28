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

import { Breadcrumbs } from '@ttn-lw/components/breadcrumbs/breadcrumbs'
import Button from '@ttn-lw/components/button-v2'
import ProfileDropdown from '@ttn-lw/components/profile-dropdown-v2'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './header-v2.styl'

const Header = ({ breadcrumbs, className, profileDropdownItems, user, ...rest }) => (
  // Const isGuest = !Boolean(user)

  <header {...rest} className={classnames(className, style.container)}>
    <Breadcrumbs breadcrumbs={breadcrumbs} />

    <div className={style.buttons}>
      <Button naked icon="add" withDropdown />
      <Button naked icon="grade" withDropdown />
      <Button naked icon="inbox" />
      <ProfileDropdown
        userName={user.name || user.ids.user_id}
        data-test-id="profile-dropdown"
        profilePicture={user.profile_picture}
      >
        {profileDropdownItems}
      </ProfileDropdown>
    </div>
  </header>
)

Header.propTypes = {
  /** A list of breadcrumb elements. */
  breadcrumbs: PropTypes.arrayOf(PropTypes.oneOfType([PropTypes.func, PropTypes.element]))
    .isRequired,
  /** The classname applied to the component. */
  className: PropTypes.string,
  profileDropdownItems: PropTypes.node.isRequired,
  /**
   * The User object, retrieved from the API. If it is `undefined`, then the
   * guest header is rendered.
   */
  user: PropTypes.user,
}

Header.defaultProps = {
  className: undefined,
  user: undefined,
}

export default Header
