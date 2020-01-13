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

import Logo from '../../containers/logo'
import Link from '../link'
import NavigationBar from '../navigation/bar'
import ProfileDropdown from '../profile-dropdown'
import Input from '../input'
import PropTypes from '../../lib/prop-types'

import style from './header.styl'

const Header = function({
  className,
  dropdownItems,
  navigationEntries,
  user,
  searchable,
  handleSearchRequest = () => null,
  ...rest
}) {
  const isGuest = !Boolean(user)

  return (
    <header {...rest} className={classnames(className, style.bar)}>
      <div className={style.left}>
        <Link to="/" className={style.logo}>
          <Logo />
        </Link>
        {!isGuest && <NavigationBar className={style.navList}>{navigationEntries}</NavigationBar>}
      </div>
      {!isGuest && (
        <div className={style.right}>
          {searchable && <Input icon="search" onEnter={handleSearchRequest} />}
          <ProfileDropdown userId={user.ids.user_id}>{dropdownItems}</ProfileDropdown>
        </div>
      )}
    </header>
  )
}

Header.propTypes = {
  /** The classname applied to the component */
  className: PropTypes.string,
  /**
   * A list of items for the dropdown
   * See `<ProfileDropdown/>`'s `items` proptypes for details
   */
  dropdownItems: ProfileDropdown.propTypes.children,
  /** A handler for when the user used the search input */
  handleSearchRequest: PropTypes.func,
  /** Child node of the navigation bar */
  navigationEntries: NavigationBar.propTypes.children,
  searchable: PropTypes.bool,
  /**
   * The User object, retrieved from the API. If it is `undefined`, then the
   * guest header is rendered
   */
  user: PropTypes.user,
}

Header.defaultProps = {
  className: undefined,
  dropdownItems: undefined,
  navigationEntries: undefined,
  handleSearchRequest: () => null,
  searchable: false,
  user: undefined,
}

export default Header
