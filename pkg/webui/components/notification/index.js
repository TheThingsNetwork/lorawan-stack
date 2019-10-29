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
import Icon from '../icon'
import PropTypes from '../../lib/prop-types'
import Details from './details'

import style from './notification.styl'

const Notification = function({
  content,
  className,
  title,
  error,
  warning,
  details,
  info,
  small,
  success,
  message,
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
    [style.withDetails]: Boolean(details),
  })

  let icon = 'info'
  if (error) {
    icon = 'error'
  } else if (warning) {
    icon = 'warning'
  }

  return (
    <div className={classname}>
      <div className={style.container}>
        <Icon className={style.icon} icon={icon} large={!small} />
        <div className={style.content}>
          {title && <Message className={style.title} content={title} component="h4" />}
          <div>
            <Message content={content} values={messageValues} />
            {action && (
              <Button
                naked
                secondary
                icon={buttonIcon}
                onClick={action}
                message={actionMessage}
                type="button"
              />
            )}
          </div>
        </div>
      </div>
      {Boolean(details) && <Details className={style.details} details={details} />}
    </div>
  )
}

Notification.propTypes = {
  action: PropTypes.func,
  actionMessage: PropTypes.message,
  buttonIcon: PropTypes.string,
  content: PropTypes.oneOfType([PropTypes.message, PropTypes.error, PropTypes.string]),
  details: PropTypes.error,
  error: PropTypes.bool,
  info: PropTypes.bool,
  message: PropTypes.bool,
  messageValues: PropTypes.object,
  small: PropTypes.bool,
  success: PropTypes.bool,
  title: PropTypes.message,
  warning: PropTypes.bool,
}

Notification.defaultProps = {
  action: undefined,
  actionMessage: undefined,
  buttonIcon: '',
  content: undefined,
  error: false,
  info: false,
  message: false,
  small: false,
  title: '',
  warning: false,
  success: false,
  messageValues: undefined,
  details: undefined,
}

export default Notification
