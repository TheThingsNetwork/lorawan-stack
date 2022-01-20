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

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import Details from './details'

import style from './notification.styl'

const Notification = ({
  content,
  className,
  title,
  error,
  warning,
  details,
  info,
  small,
  success,
  messageValues = {},
  children,
  convertBackticks,
  'data-test-id': dataTestId,
}) => {
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
    <div className={classname} data-test-id={dataTestId}>
      <div className={style.container}>
        <Icon className={style.icon} icon={icon} large={!small} />
        <div className={style.content}>
          {title && <Message className={style.title} content={title} component="h4" />}
          <div>
            <Message
              content={content}
              values={messageValues}
              firstToUpper
              convertBackticks={Boolean(error) || convertBackticks}
            />
            {children}
          </div>
        </div>
      </div>
      {Boolean(details) && <Details className={style.details} details={details} />}
    </div>
  )
}

Notification.propTypes = {
  children: PropTypes.node,
  className: PropTypes.string,
  content: PropTypes.oneOfType([PropTypes.message, PropTypes.error, PropTypes.string]),
  convertBackticks: PropTypes.bool,
  'data-test-id': PropTypes.string,
  details: PropTypes.error,
  error: PropTypes.bool,
  info: PropTypes.bool,
  messageValues: PropTypes.shape({}),
  small: PropTypes.bool,
  success: PropTypes.bool,
  title: PropTypes.message,
  warning: PropTypes.bool,
}

Notification.defaultProps = {
  children: undefined,
  className: undefined,
  content: undefined,
  'data-test-id': 'notification',
  error: false,
  info: false,
  small: false,
  title: '',
  warning: false,
  success: false,
  messageValues: undefined,
  details: undefined,
  convertBackticks: false,
}

export default Notification
