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

import React from 'react'
import classnames from 'classnames'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import Icon from '../icon'

import style from './status-label.styl'

const StatusLabel = ({ success, warning, error, info, content }) => {
  const statusClassName = classnames(style.label, {
    'c-bg-success-light c-text-success-bold': success,
    'c-bg-warning-light c-text-warning-bold': warning,
    'c-bg-error-light c-text-error-bold': error,
    'c-bg-info-light c-text-info-bold': info,
  })

  const labelIcon = success ? 'check_circle' : warning ? 'warning' : error ? 'error' : 'info'

  return (
    <div className={statusClassName}>
      <Icon icon={labelIcon} className={style.labelIcon} />
      <Message content={content} className={style.labelContent} />
    </div>
  )
}

StatusLabel.propTypes = {
  content: PropTypes.message.isRequired,
  error: PropTypes.bool,
  info: PropTypes.bool,
  success: PropTypes.bool,
  warning: PropTypes.bool,
}

StatusLabel.defaultProps = {
  success: false,
  warning: false,
  error: false,
  info: false,
}

export default StatusLabel
