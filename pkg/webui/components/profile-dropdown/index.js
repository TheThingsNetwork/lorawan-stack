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

import ProfilePicture from '@ttn-lw/components/profile-picture'
import Button from '@ttn-lw/components/button'

import PropTypes from '@ttn-lw/lib/prop-types'

import styles from './profile-dropdown.styl'

const ProfileDropdown = props => {
  const { brandLogo, className, children, profilePicture, darkTheme, ...rest } = props

  return (
    <div className={styles.container}>
      <Button
        secondary
        className={classnames(styles.profileButton, className, 'pr-0')}
        dropdownItems={children}
        dropdownPosition="below left"
        {...rest}
      >
        <div className="d-flex gap-cs-xs al-center">
          {brandLogo && <img {...brandLogo} className={styles.brandLogo} />}
          <ProfilePicture
            className={styles.profilePicture}
            darkTheme={darkTheme}
            profilePicture={profilePicture}
          />
        </div>
      </Button>
    </div>
  )
}

ProfileDropdown.propTypes = {
  brandLogo: PropTypes.shape({
    src: PropTypes.string.isRequired,
    alt: PropTypes.string.isRequired,
  }),
  /**
   * A list of items for the dropdown component. See `<Dropdown />`'s `items`
   * proptypes for details.
   */
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  /** The dark theme. */
  darkTheme: PropTypes.bool,
  /** The profile picture of the current user. */
  profilePicture: PropTypes.profilePicture,
}

ProfileDropdown.defaultProps = {
  brandLogo: undefined,
  className: undefined,
  profilePicture: undefined,
  darkTheme: false,
}

export default ProfileDropdown
