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

import React, { useCallback, useRef, useState } from 'react'
import classnames from 'classnames'

import Icon from '@ttn-lw/components/icon'
import Dropdown from '@ttn-lw/components/dropdown'
import ProfilePicture from '@ttn-lw/components/profile-picture'

import PropTypes from '@ttn-lw/lib/prop-types'

import styles from './profile-dropdown.styl'

const ProfileDropdown = props => {
  const [expanded, setExpanded] = useState(false)
  const node = useRef(null)
  const { userName, className, children, profilePicture, ...rest } = props

  const handleClickOutside = useCallback(e => {
    if (node.current && !node.current.contains(e.target)) {
      setExpanded(false)
    }
  }, [])

  const toggleDropdown = useCallback(() => {
    setExpanded(oldExpanded => {
      const newState = !oldExpanded
      if (newState) document.addEventListener('mousedown', handleClickOutside)
      else document.removeEventListener('mousedown', handleClickOutside)
      return newState
    })
  }, [handleClickOutside])

  return (
    <div
      className={classnames(styles.container, className)}
      onClick={toggleDropdown}
      ref={node}
      tabIndex="0"
      role="button"
      {...rest}
    >
      <ProfilePicture profilePicture={profilePicture} className={styles.profilePicture} />
      <span className={styles.id}>{userName}</span>
      <Icon className={styles.dropdownIcon} icon={expanded ? 'arrow_drop_up' : 'arrow_drop_down'} />
      {expanded && <Dropdown className={styles.dropdown}>{children}</Dropdown>}
    </div>
  )
}

ProfileDropdown.propTypes = {
  /**
   * A list of items for the dropdown component. See `<Dropdown />`'s `items`
   * proptypes for details.
   */
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  /** The profile picture of the current user. */
  profilePicture: PropTypes.profilePicture,
  /** The name/id of the current user. */
  userName: PropTypes.string.isRequired,
}

ProfileDropdown.defaultProps = {
  className: undefined,
  profilePicture: undefined,
}

export default ProfileDropdown
