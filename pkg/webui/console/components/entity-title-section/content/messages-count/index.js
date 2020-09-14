// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import { FormattedNumber, useIntl } from 'react-intl'

import Icon from '@ttn-lw/components/icon'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './messages-count.styl'

const MessagesCount = props => {
  const { className, icon, value, tooltipMessage, iconClassName } = props
  const { formatMessage } = useIntl()

  return (
    <div title={formatMessage(tooltipMessage)} className={classnames(style.container, className)}>
      <Icon className={iconClassName} icon={icon} nudgeUp />
      {typeof value === 'number' ? <FormattedNumber value={value} /> : value}
    </div>
  )
}

MessagesCount.propTypes = {
  className: PropTypes.string,
  icon: PropTypes.string.isRequired,
  iconClassName: PropTypes.string,
  tooltipMessage: PropTypes.message.isRequired,
  value: PropTypes.oneOfType([PropTypes.number, PropTypes.string]).isRequired,
}

MessagesCount.defaultProps = {
  className: undefined,
  iconClassName: undefined,
}

export default MessagesCount
