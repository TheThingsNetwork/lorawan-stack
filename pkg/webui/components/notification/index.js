// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

import Message from '../message'
import ErrorMessage from '../error-message'
import Icon from '../icon'

import PropTypes from '../../lib/prop-types'
import style from './notification.styl'

const Notification = function ({
  className,
  error,
  warning,
  info,
  small,
  message,
}) {

  const classname = classnames(style.notification, className, {
    [style.error]: error,
    [style.warning]: warning,
    [style.info]: info,
    [style.small]: small,
  })

  let icon = 'info'
  if (error) {
    icon = 'error'
  } else if (warning) {
    icon = 'warning'
  }

  const content = message || error || warning || info
  const Component = error ? ErrorMessage : Message

  return (
    <div className={classname}>
      <Icon icon={icon} /><span><Component content={content} /></span>
    </div>
  )
}

Notification.propTypes = {
  message: PropTypes.message,
  error: PropTypes.error,
  warning: PropTypes.message,
  info: PropTypes.message,
  small: PropTypes.bool,
}

export default Notification
