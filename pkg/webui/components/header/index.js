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

import React, { useState, useCallback } from 'react'
import classnames from 'classnames'

import hamburgerMenuNormal from '@assets/misc/hamburger-menu-normal.svg'
import hamburgerMenuClose from '@assets/misc/hamburger-menu-close.svg'

import Icon from '@ttn-lw/components/icon'
import NavigationBar from '@ttn-lw/components/navigation/bar'
import ProfileDropdown from '@ttn-lw/components/profile-dropdown'
import MobileMenu from '@ttn-lw/components/mobile-menu'
import Input from '@ttn-lw/components/input'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './header.styl'

const Header = function({
  className,
  dropdownItems,
  navigationEntries,
  user,
  searchable,
  logo,
  mobileDropdownItems,
  onLogout,
  onSearchRequest,
  ...rest
}) {
  const isGuest = !Boolean(user)

  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const handleMobileMenuClick = useCallback(() => {
    setMobileMenuOpen(!mobileMenuOpen)
  }, [mobileMenuOpen])

  const handleMobileMenuItemsClick = useCallback(() => {
    setMobileMenuOpen(false)
  }, [])

  const classNames = classnames(className, style.container, {
    [style.mobileMenuOpen]: mobileMenuOpen,
  })

  const hamburgerGraphic = mobileMenuOpen ? hamburgerMenuClose : hamburgerMenuNormal

  return (
    <header {...rest} className={classNames}>
      <div className={style.bar}>
        <div className={style.left}>
          {logo}
          {!isGuest && <NavigationBar className={style.navList}>{navigationEntries}</NavigationBar>}
        </div>
        {!isGuest && (
          <div className={style.right}>
            {searchable && <Input icon="search" onEnter={onSearchRequest} />}
            <ProfileDropdown
              className={style.profileDropdown}
              userName={user.name || user.ids.user_id}
              data-test-id="profile-dropdown"
            >
              {dropdownItems}
            </ProfileDropdown>
            <button onClick={handleMobileMenuClick} className={style.mobileMenu}>
              <Icon className={style.preloadIcons} icon="." />
              <div className={style.hamburger}>
                <img src={hamburgerGraphic} alt="Open Mobile Menu" />
              </div>
            </button>
          </div>
        )}
      </div>
      {mobileMenuOpen && (
        <MobileMenu onItemsClick={handleMobileMenuItemsClick} onLogout={onLogout} user={user}>
          {mobileDropdownItems}
        </MobileMenu>
      )}
    </header>
  )
}

Header.propTypes = {
  /** The classname applied to the component. */
  className: PropTypes.string,
  /** The child node of the dropdown component. */
  dropdownItems: ProfileDropdown.propTypes.children,
  /** The logo component. */
  logo: PropTypes.node.isRequired,
  /** The child node of the mobile dropdown. */
  mobileDropdownItems: PropTypes.node.isRequired,
  /** The Child node of the navigation bar. */
  navigationEntries: NavigationBar.propTypes.children,
  /** A handler for when the user used the search input. */
  onLogout: PropTypes.func.isRequired,
  /** Handler of the search function. */
  onSearchRequest: PropTypes.func,
  /* A flag indicating whether the header has a search input. */
  searchable: PropTypes.bool,
  /**
   * The User object, retrieved from the API. If it is `undefined`, then the
   * guest header is rendered.
   */
  user: PropTypes.user,
}

Header.defaultProps = {
  className: undefined,
  dropdownItems: undefined,
  navigationEntries: undefined,
  onSearchRequest: () => null,
  searchable: false,
  user: undefined,
}

export default Header
