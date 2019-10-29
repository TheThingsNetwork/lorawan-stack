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
import { toast as t } from 'react-toastify'

import Notification from '../notification'

import style from './toast.styl'

const createToast = function() {
  const queue = []
  let toastId = null
  let firstDispatched = false

  const show = function(toastOptions) {
    const { INFO, SUCCESS, ERROR, WARNING, DEFAULT } = toast.types
    const { title, message, type = DEFAULT, ...rest } = toastOptions

    toastId = t(
      <Notification
        className={style.notification}
        small
        title={title}
        content={message}
        success={type === SUCCESS}
        info={type === INFO}
        error={type === ERROR}
        warning={type === WARNING}
        message={type === DEFAULT}
      />,
      {
        onClose: () => next(),
        ...rest,
      },
    )
  }

  const next = function() {
    const hasNext = queue.length > 0

    if (!hasNext) {
      firstDispatched = false
    } else if (hasNext && !t.isActive(toastId)) {
      show(queue.shift())
    }
  }

  const toast = function(options) {
    queue.push(options)

    if (!firstDispatched) {
      firstDispatched = true
      next()
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
