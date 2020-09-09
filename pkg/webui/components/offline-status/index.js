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

import React, { useEffect, useRef } from 'react'
import { defineMessages } from 'react-intl'
import classnames from 'classnames'

import toast from '@ttn-lw/components/toast'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './offline.styl'

const m = defineMessages({
  offline: 'The application went offline',
  online: 'The application is back online',
})

const handleMessage = (message, type) => {
  // Don't show a toast when the tab is not in focus
  // to prevent flooding the toast queue.
  if (document.hidden) {
    return
  }

  toast({
    message,
    type,
  })
}

const OfflineStatus = ({ showOfflineOnly, showWarnings, isOnline }) => {
  const initialUpdate = useRef(true)
  useEffect(() => {
    if (initialUpdate.current) {
      initialUpdate.current = false
      return
    }
    if (showWarnings && isOnline) {
      handleMessage(m.online, toast.types.INFO)
    } else if (showWarnings && !isOnline) {
      handleMessage(m.offline, toast.types.ERROR)
    }
  }, [showWarnings, isOnline])

  if (showOfflineOnly && isOnline) {
    return null
  }

  return (
    <span className={classnames(style.status, { [style.online]: isOnline })}>
      <Icon className={style.icon} icon={isOnline ? 'info' : 'error'} />
      <Message content={isOnline ? sharedMessages.online : sharedMessages.offline} />
    </span>
  )
}

OfflineStatus.propTypes = {
  isOnline: PropTypes.bool.isRequired,
  showOfflineOnly: PropTypes.bool,
  showWarnings: PropTypes.bool,
}

OfflineStatus.defaultProps = {
  showOfflineOnly: false,
  showWarnings: false,
}

export default OfflineStatus
