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

import Button from '../button'
import Message from '../../lib/components/message'
import ErrorMessage from '../../lib/components/error-message'
import Icon from '../icon'

import PropTypes from '../../lib/prop-types'
import style from './notification.styl'

const Notification = function({
  className,
  title,
  error,
  warning,
  info,
  small,
  message,
  success,
  messageValues = {},
  action,
  actionMessage,
  buttonIcon,
}) {
  const classname = classnames(style.notification, className, {
    [style.error]: error,
    [style.warning]: warning,
    [style.info]: info,
    [style.small]: small,
    [style.success]: success,
  })

  let icon = 'info'
  if (error) {
    icon = 'error'
  } else if (warning) {
    icon = 'warning'
  }

  const content = message || error || warning || info || success
  const Component = error ? ErrorMessage : Message

  return (
    <div className={classname}>
      <Icon className={style.icon} icon={icon} large={!small} />
      <div className={style.content}>
        {title && <Message className={style.title} content={title} component="h4" />}
        <div>
          <Component content={content} values={messageValues} />
          {action && (
            <Button naked secondary icon={buttonIcon} onClick={action} message={actionMessage} />
          )}
        </div>
      </div>
    </div>
  )
}

Notification.propTypes = {
  action: PropTypes.func,
  actionMessage: PropTypes.message,
  buttonIcon: PropTypes.string,
  error: PropTypes.error,
  info: PropTypes.message,
  message: PropTypes.message,
  messageValues: PropTypes.object,
  small: PropTypes.bool,
  success: PropTypes.message,
  title: PropTypes.message,
  warning: PropTypes.message,
}

Notification.defaultProps = {
  action: () => null,
  actionMessage: undefined,
  buttonIcon: '',
}

export default Notification
