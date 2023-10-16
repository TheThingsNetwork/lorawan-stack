// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useState, useEffect } from 'react'
import { defineMessages } from 'react-intl'

import ErrorMessage from '@ttn-lw/lib/components/error-message'

import PropTypes from '@ttn-lw/lib/prop-types'

import Button from '../button'

import Capture from './input/capture'

import style from './qr.styl'

const m = defineMessages({
  permissionDeniedError: 'Permission Denied: Please allow access to your camera or upload a photo',
  fetchingPermission: 'Please set camera permissions',
})

const RequirePermission = props => {
  const { children, onRead, useCapture, setCapture, videoError } = props
  const [allow, setAllow] = useState(false)
  const [permission, setPermission] = useState(undefined)

  const handleRead = useCallback(
    (media, width, height) => {
      onRead(media, width, height)
    },
    [onRead],
  )

  const handlePermissionState = useCallback(
    state => {
      switch (state) {
        case 'granted':
        case 'prompt':
          setAllow(true)
          setCapture(false)
          break
        case 'denied':
          setAllow(false)
          break
      }
    },
    [setCapture],
  )

  const handlePermissionChange = useCallback(
    event => {
      if (event.target instanceof PermissionStatus) {
        handlePermissionState(event.target.state)
      }
    },
    [handlePermissionState],
  )

  const getCameraPermission = useCallback(async () => {
    try {
      const permission = await navigator.permissions.query({ name: 'camera' })
      setPermission(permission)
    } catch (error) {
      if (error instanceof TypeError) {
        // Always allow for browsers that do not support the permissions API for camera.
        setAllow(true)
      } else {
        throw error
      }
    }
  }, [])

  const handleUseCapture = useCallback(() => {
    setCapture(true)
  }, [setCapture])

  // Lookup camera permission based on the Permissions API
  // store obj in component state and display either video
  // or video error message components.
  // Add event listener to update displayed component
  // based on external change to permission without page refresh.
  useEffect(() => {
    if (permission) {
      handlePermissionState(permission.state)
      permission.addEventListener('change', handlePermissionChange)
    } else {
      getCameraPermission()
    }
    return () => {
      if (permission) {
        permission.removeEventListener('change', handlePermissionChange)
      }
    }
  }, [getCameraPermission, handlePermissionChange, handlePermissionState, permission])

  if (useCapture) {
    return <Capture onRead={handleRead} />
  }

  if (!allow || videoError) {
    return (
      <div className={style.captureWrapper}>
        <ErrorMessage style={{ color: '#fff' }} content={m.permissionDeniedError} />
        <br />
        <Button className="mt-cs-m" onClick={handleUseCapture} message={m.uploadImage} />
      </div>
    )
  }

  return children
}

RequirePermission.propTypes = {
  children: PropTypes.node.isRequired,
  onRead: PropTypes.func.isRequired,
  setCapture: PropTypes.func.isRequired,
  useCapture: PropTypes.bool.isRequired,
  videoError: PropTypes.bool,
}

RequirePermission.defaultProps = {
  videoError: false,
}

export default RequirePermission
