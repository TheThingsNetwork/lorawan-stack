// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import React, { useRef, useCallback, useEffect } from 'react'
import classnames from 'classnames'
import { useSelector } from 'react-redux'

import missingProfilePicture from '@assets/img/placeholder/missing-profile-picture.svg'
import missingProfilePictureWhite from '@assets/img/placeholder/missing-profile-picture-white.svg'

import PropTypes from '@ttn-lw/lib/prop-types'
import {
  getClosestProfilePictureBySize,
  isValidProfilePictureObject,
} from '@ttn-lw/lib/selectors/profile-picture'

import { selectConsolePreferences } from '@console/store/selectors/user-preferences'

import styles from './profile-picture.styl'

const ProfilePicture = ({ profilePicture, className, size }) => {
  const imageRef = useRef()
  const consolePreferences = useSelector(selectConsolePreferences)
  const darkTheme =
    consolePreferences.console_theme === 'CONSOLE_THEME_DARK' ||
    (consolePreferences.console_theme === 'CONSOLE_THEME_SYSTEM' &&
      window.matchMedia('(prefers-color-scheme: dark)').matches)

  const handleImageError = useCallback(
    error => {
      error.target.src = darkTheme ? missingProfilePictureWhite : missingProfilePicture
    },
    [darkTheme],
  )

  useEffect(() => {
    if (imageRef.current) {
      imageRef.current.src = darkTheme ? missingProfilePictureWhite : missingProfilePicture
    }
  }, [darkTheme])

  return (
    <div className={classnames(className, styles.container)}>
      <img
        onError={handleImageError}
        src={
          isValidProfilePictureObject(profilePicture)
            ? getClosestProfilePictureBySize(profilePicture, size)
            : darkTheme
              ? missingProfilePictureWhite
              : missingProfilePicture
        }
        alt="Profile picture"
        ref={imageRef}
      />
    </div>
  )
}

ProfilePicture.propTypes = {
  className: PropTypes.string,
  profilePicture: PropTypes.profilePicture,
  size: PropTypes.number,
}

ProfilePicture.defaultProps = {
  profilePicture: undefined,
  className: undefined,
  size: 128,
}

export default ProfilePicture
