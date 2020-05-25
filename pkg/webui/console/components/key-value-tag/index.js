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
import { FormattedNumber, injectIntl } from 'react-intl'

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './key-value-tag.styl'

const KeyValueTag = ({ icon, keyMessage, value }) => {
  return (
    <div className={style.container}>
      <Icon className={style.icon} icon={icon} nudgeUp />
      <React.Fragment>
        <span className={style.value}>{value} </span>
        <Message
          className={style.message}
          content={keyMessage}
          values={{ count: typeof value === 'number' ? value : undefined }}
        />
      </React.Fragment>
    </div>
  )
}

KeyValueTag.propTypes = {
  icon: PropTypes.string.isRequired,
  keyMessage: PropTypes.message.isRequired,
  value: PropTypes.node.isRequired,
}

const IconValueTag = ({ icon, value, tooltipMessage, className, iconClassName, intl }) => (
  <div title={intl.formatMessage(tooltipMessage)} className={className}>
    <Icon className={iconClassName} icon={icon} nudgeUp />
    {typeof value === 'number' ? <FormattedNumber value={value} /> : value}
  </div>
)

IconValueTag.propTypes = {
  className: PropTypes.string,
  icon: PropTypes.string.isRequired,
  iconClassName: PropTypes.string,
  intl: PropTypes.shape({
    formatMessage: PropTypes.func.isRequired,
  }).isRequired,
  tooltipMessage: PropTypes.message.isRequired,
  value: PropTypes.oneOfType([PropTypes.number, PropTypes.string]).isRequired,
}

IconValueTag.defaultProps = {
  className: undefined,
  iconClassName: undefined,
}

const IntlIconValueTag = injectIntl(IconValueTag)

export { KeyValueTag as default, IntlIconValueTag as IconValueTag }
