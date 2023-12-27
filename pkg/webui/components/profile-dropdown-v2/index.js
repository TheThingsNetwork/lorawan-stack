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
import Dropdown from '@ttn-lw/components/dropdown-v2'
import ProfilePicture from '@ttn-lw/components/profile-picture'
import Button from '@ttn-lw/components/button'
import style from '@ttn-lw/components/button-v2/button.styl'

import PropTypes from '@ttn-lw/lib/prop-types'

import styles from './profile-dropdown-v2.styl'

const ProfileDropdown = props => {
  const [expanded, setExpanded] = useState(false)
  const node = useRef(null)
  const { brandLogo, className, children, profilePicture, ...rest } = props

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
    <Button
      secondary
      className={classnames(styles.container, className)}
      onClick={toggleDropdown}
      ref={node}
      {...rest}
    >
      <div className="d-flex gap-cs-xs al-center">
        {brandLogo && <img {...brandLogo} className={styles.brandLogo} />}
        <ProfilePicture className={styles.profilePicture} profilePicture={profilePicture} />
      </div>
      <Icon
        className={classnames(style.arrowIcon, {
          [style['arrow-icon-expanded']]: expanded,
        })}
        icon="expand_more"
      />
      <Dropdown open={expanded}>{children}</Dropdown>
    </Button>
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
  /** The profile picture of the current user. */
  profilePicture: PropTypes.profilePicture,
}

ProfileDropdown.defaultProps = {
  brandLogo: undefined,
  className: undefined,
  profilePicture: undefined,
}

export default ProfileDropdown
