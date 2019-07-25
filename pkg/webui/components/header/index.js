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

import Logo from '../logo'
import NavigationBar from '../navigation/bar'
import ProfileDropdown from '../profile-dropdown'
import Input from '../input'
import PropTypes from '../../lib/prop-types'
import sharedMessages from '../../lib/shared-messages'

import styles from './header.styl'

const defaultNavigationEntries = [
  {
    title: sharedMessages.overview,
    icon: 'overview',
    path: '/console',
    exact: true,
  },
  {
    title: sharedMessages.applications,
    icon: 'application',
    path: '/console/applications',
  },
  {
    title: sharedMessages.gateways,
    icon: 'gateway',
    path: '/console/gateways',
  },
]

const defaultDropdownItems = handleLogout => [
  {
    title: sharedMessages.logout,
    icon: 'power_settings_new',
    action: handleLogout,
  },
]

const Header = function ({
  className,
  handleLogout = () => null,
  dropdownItems = defaultDropdownItems(handleLogout),
  navigationEntries = defaultNavigationEntries,
  user,
  searchable,
  handleSearchRequest = () => null,
  anchored = false,
  ...rest
}) {
  const isGuest = !Boolean(user)

  return (
    <header {...rest} className={classnames(className, styles.bar)}>
      {
        isGuest ? (
          <div className={styles.left}>
            <div className={styles.logo}><Logo /></div>
          </div>
        ) : (
          <React.Fragment>
            <div className={styles.left}>
              <div className={styles.logo}><Logo /></div>
              <NavigationBar
                className={styles.navList}
                entries={navigationEntries}
                anchored={anchored}
              />
            </div>
            <div className={styles.right}>
              { searchable && <Input icon="search" onEnter={handleSearchRequest} /> }
              <ProfileDropdown
                dropdownItems={dropdownItems || defaultDropdownItems}
                userId={user.ids.user_id}
                anchored={anchored}
              />
            </div>
          </React.Fragment>
        )
      }
    </header>
  )
}

Header.propTypes = {
  /**
  * The User object, retrieved from the API. If it is
  * `undefined`, then the guest header is rendered.
  */
  user: PropTypes.object,
  /**
  * A list of items for the dropdown
  * @param {(string|Object)} title - The title to be displayed
  * @param {string} icon - The icon name to be displayed next to the title
  * @param {string} path - The path for a navigation tab
  * @param {function} action - Alternatively, the function to be called on click
  */
  dropdownItems: PropTypes.arrayOf(PropTypes.shape({
    title: PropTypes.message.isRequired,
    icon: PropTypes.string,
    path: PropTypes.string.isRequired,
    action: PropTypes.func,
  })),
  /**
   * A list of navigation bar entries.
   * @param {(string|Object)} title - The title to be displayed
   * @param {string} icon - The icon name to be displayed next to the title
   * @param {string} path -  The path for a navigation tab
   * @param {boolean} exact - Flag identifying whether the path should be matched exactly
   */
  navigationEntries: PropTypes.arrayOf(PropTypes.shape({
    path: PropTypes.string.isRequired,
    title: PropTypes.message.isRequired,
    action: PropTypes.func,
    icon: PropTypes.string,
  })),
  /** Flag identifying whether links should be rendered as plain anchor link */
  anchored: PropTypes.bool,
  /**
  * A handler for when the user clicks the logout button
  */
  handleLogout: PropTypes.func,
  /**
  * A handler for when the user used the search input
  */
  handleSearchRequest: PropTypes.func,
  /** A flag identifying whether the header should display the search input */
  searchable: PropTypes.bool,
}

export default Header
