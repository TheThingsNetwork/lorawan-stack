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

import Button from '@ttn-lw/components/button'
import ProfileDropdown from '@ttn-lw/components/profile-dropdown'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './header.styl'

const Header = ({
  brandLogo,
  Logo,
  className,
  addDropdownItems,
  starDropdownItems,
  profileDropdownItems,
  notificationsDropdownItems,
  user,
  onMenuClick,
  showNotificationDot,
  ...rest
}) => (
  <header {...rest} className={classnames(className, style.container)}>
    <div className={classnames('breadcrumbs', 'm:d-none')} />
    <div className="d-none m:d-flex al-center gap-cs-xs">
      <Button secondary icon="menu" onClick={onMenuClick} />
      <Logo className={style.logo} />
    </div>

    <div className="d-flex al-center gap-cs-xs">
      <Button
        secondary
        icon="add_box"
        dropdownItems={addDropdownItems}
        dropdownPosition="below left"
      />
      <Button
        secondary
        icon="bookmark"
        dropdownItems={starDropdownItems}
        dropdownPosition="below left"
        className="xs:d-none"
      />
      <Button
        secondary
        icon="inbox"
        dropdownItems={notificationsDropdownItems}
        dropdownClassName={style.notificationsDropdown}
        dropdownPosition="below left"
        withAlert={showNotificationDot}
      />
      <ProfileDropdown
        brandLogo={brandLogo}
        data-test-id="profile-dropdown"
        profilePicture={user?.profile_picture}
      >
        {profileDropdownItems}
      </ProfileDropdown>
    </div>
  </header>
)

const imgPropType = PropTypes.shape({
  src: PropTypes.string.isRequired,
  alt: PropTypes.string.isRequired,
})

Header.propTypes = {
  Logo: PropTypes.elementType.isRequired,
  /** The dropdown items when the add button is clicked. */
  addDropdownItems: PropTypes.node.isRequired,
  brandLogo: imgPropType,
  /** The classname applied to the component. */
  className: PropTypes.string,
  /** The dropdown items when the notifications button is clicked. */
  notificationsDropdownItems: PropTypes.node.isRequired,
  /** A handler for when the menu button is clicked. */
  onMenuClick: PropTypes.func.isRequired,
  /** The dropdown items when the profile button is clicked. */
  profileDropdownItems: PropTypes.node.isRequired,
  /** Whether to show a notification dot. */
  showNotificationDot: PropTypes.bool,
  /** The dropdown items when the star button is clicked. */
  starDropdownItems: PropTypes.node.isRequired,
  /**
   * The User object, retrieved from the API. If it is `undefined`, then the
   * guest header is rendered.
   */
  user: PropTypes.user,
}

Header.defaultProps = {
  className: undefined,
  user: undefined,
  brandLogo: undefined,
  showNotificationDot: false,
}

export default Header
