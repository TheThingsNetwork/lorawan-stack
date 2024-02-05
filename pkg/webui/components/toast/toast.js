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
import { toast as t, cssTransition } from 'react-toastify'

import Notification from '@ttn-lw/components/notification'

import diff from '@ttn-lw/lib/diff'

import style from './toast.styl'

const createToast = () => {
  let lastMessage = undefined
  let lastMessageGroup = undefined
  let toastId = null
  let currentRender

  const show = toastOptions => {
    if (!currentRender) {
      return
    }

    toastId = t(currentRender, toastOptions)
  }

  const toast = options => {
    // Prevent flooding of identical messages (if wished).
    if (
      options.preventConsecutive &&
      lastMessage &&
      Object.keys(diff(lastMessage, options)).length === 0
    ) {
      return
    }

    const { INFO, SUCCESS, ERROR, WARNING, DEFAULT } = toast.types
    const {
      title,
      message,
      messageValues = {},
      type = DEFAULT,
      messageGroup,
      ...toastOptions
    } = options
    let autoClose = toastOptions.autoClose

    if (!autoClose) {
      let messageLength =
        typeof message === 'string'
          ? message.length
          : typeof message === 'object' && message.defaultMessage
            ? message.defaultMessage.length
            : 0
      if (title) {
        messageLength +=
          typeof title === 'string'
            ? title.length
            : typeof title === 'object' && title.defaultMessage
              ? title.defaultMessage.length
              : 0
      }
      // Calculate the reading time to use as `autoClose` duration.
      autoClose = Math.min(12000, Math.max(5000, messageLength * 150))
    }

    currentRender = (
      <Notification
        className={style.notification}
        small
        title={title}
        content={message}
        messageValues={messageValues}
        success={type === SUCCESS}
        info={type === INFO}
        error={type === ERROR}
        warning={type === WARNING}
        data-test-id="toast-notification"
      />
    )

    // For messages of the same message group, update the card rather than
    // queuing up more messages.
    if (t.isActive(toastId) && messageGroup === lastMessageGroup) {
      t.update(toastId, {
        ...toastOptions,
        render: currentRender,
        autoClose,
        transition: cssTransition({ enter: style.beat, exit: style.slideOutRight }),
      })
    } else {
      show({ autoClose, ...toastOptions })
      lastMessage = options
      lastMessageGroup = messageGroup
    }
  }

  toast.types = {
    INFO: 'info',
    SUCCESS: 'success',
    ERROR: 'error',
    WARNING: 'warning',
    DEFAULT: 'default',
  }

  return toast
}

export default createToast
