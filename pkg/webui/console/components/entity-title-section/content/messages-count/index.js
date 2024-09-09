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
import { FormattedNumber } from 'react-intl'

import Icon, { IconHelp } from '@ttn-lw/components/icon'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './messages-count.styl'

const MessagesCount = React.forwardRef((props, ref) => {
  const { className, icon, value, iconClassName, helpTooltip } = props

  return (
    <div ref={ref} className={classnames(style.container, className)}>
      <Icon className={iconClassName} icon={icon} nudgeUp />
      {typeof value === 'number' ? <FormattedNumber value={value} /> : value}
      {helpTooltip && (
        <Icon className="c-text-neutral-light ml-cs-xxs" icon={IconHelp} nudgeUp small />
      )}
    </div>
  )
})

MessagesCount.propTypes = {
  className: PropTypes.string,
  helpTooltip: PropTypes.bool,
  icon: PropTypes.icon.isRequired,
  iconClassName: PropTypes.string,
  value: PropTypes.node.isRequired,
}

MessagesCount.defaultProps = {
  className: undefined,
  iconClassName: undefined,
  helpTooltip: false,
}

export default MessagesCount
