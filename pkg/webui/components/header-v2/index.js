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

import React, { useRef } from 'react'
import classnames from 'classnames'

import { Breadcrumbs } from '@ttn-lw/components/breadcrumbs/breadcrumbs'
import Button from '@ttn-lw/components/button-v2'
import ProfileDropdown from '@ttn-lw/components/profile-dropdown-v2'
import Link from '@ttn-lw/components/link'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './header-v2.styl'

const Header = ({
  brandLogo,
  logo,
  breadcrumbs,
  className,
  addDropdownItems,
  starDropdownItems,
  profileDropdownItems,
  user,
  onMenuClick,
  ...rest
}) => {
  const addRef = useRef(null)
  const starRef = useRef(null)

  // Const isGuest = !Boolean(user)

  return (
    <header {...rest} className={classnames(className, style.container)}>
      <Breadcrumbs className="s:d-none" breadcrumbs={breadcrumbs} />
      <div className="d-none s:d-flex al-center gap-cs-xs">
        <Button secondary icon="menu" onClick={onMenuClick} />
        <Link to="/" className="d-flex">
          <img {...logo} className={style.logo} />
        </Link>
      </div>

      <div className="d-flex al-center gap-cs-xs">
        <Button secondary icon="add" dropdownItems={addDropdownItems} ref={addRef} />
        <Button
          secondary
          icon="grade"
          dropdownItems={starDropdownItems}
          ref={starRef}
          className="s:d-none"
        />
        <Button secondary icon="inbox" />
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
}

const imgPropType = PropTypes.shape({
  src: PropTypes.string.isRequired,
  alt: PropTypes.string.isRequired,
})

Header.propTypes = {
  /** The dropdown items when the add button is clicked. */
  addDropdownItems: PropTypes.node.isRequired,
  brandLogo: imgPropType,
  /** A list of breadcrumb elements. */
  breadcrumbs: PropTypes.arrayOf(PropTypes.oneOfType([PropTypes.func, PropTypes.element]))
    .isRequired,
  /** The classname applied to the component. */
  className: PropTypes.string,
  logo: imgPropType.isRequired,
  /** A handler for when the menu button is clicked. */
  onMenuClick: PropTypes.func.isRequired,
  /** The dropdown items when the profile button is clicked. */
  profileDropdownItems: PropTypes.node.isRequired,
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
}

export default Header
