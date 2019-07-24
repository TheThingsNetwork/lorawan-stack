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

import styles from './header.styl'

const Header = function ({
  className,
  handleLogout = () => null,
  dropdownItems,
  navigationEntries,
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
                dropdownItems={dropdownItems}
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
    path: PropTypes.string,
    action: PropTypes.func,
  })).isRequired,
  /**
   * A list of navigation bar entries.
   * @param {(string|Object)} title - The title to be displayed
   * @param {string} icon - The icon name to be displayed next to the title
   * @param {string} path -  The path for a navigation tab
   * @param {boolean} exact - Flag identifying whether the path should be matched exactly
   */
  entries: PropTypes.arrayOf(PropTypes.shape({
    path: PropTypes.string.isRequired,
    title: PropTypes.message.isRequired,
    icon: PropTypes.string,
    exact: PropTypes.bool,
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
