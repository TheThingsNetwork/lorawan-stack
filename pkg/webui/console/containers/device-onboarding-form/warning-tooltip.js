// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import { defineMessages } from 'react-intl'
import { isNumber } from 'lodash'

import Tooltip from '@ttn-lw/components/tooltip'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

const isEqual = (current, desired) => {
  if (isNumber(current) || isNumber(desired)) {
    return Number(current) === Number(desired)
  }

  return current === desired
}

const m = defineMessages({
  desiredDescription:
    'The network will use a different <i>desired</i> value of <code>{value}</code> for this property.',
  sessionDescription:
    'An ABP device is personalized with a session and MAC settings. These MAC settings are considered the current parameters and must match exactly the settings entered here. The Network Server uses desired parameters to change the MAC state with LoRaWAN MAC commands to the desired state. You can use the General Settings page to update the desired setting after you registered the end device.',
})

const Content = props => {
  const { desired } = props

  return (
    <div>
      <Message
        content={m.desiredDescription}
        values={{
          value: desired,
          i: txt => <i>{txt}</i>,
          code: txt => <code>{txt}</code>,
        }}
      />
      <Message content={m.sessionDescription} component="p" />
    </div>
  )
}

Content.propTypes = {
  desired: PropTypes.oneOfType([PropTypes.string, PropTypes.number]).isRequired,
}

const WarningTooltip = props => {
  const { desiredValue, currentValue } = props

  const hasDesiredValue = typeof desiredValue !== 'undefined'
  const hasCurrentValue = typeof currentValue !== 'undefined'

  if (hasDesiredValue && hasCurrentValue && !isEqual(currentValue, desiredValue)) {
    return (
      <Tooltip placement="bottom-start" interactive content={<Content desired={desiredValue} />}>
        <Icon icon="warning" small className="c-bg-warning-normal" nudgeUp />
      </Tooltip>
    )
  }

  return null
}

WarningTooltip.propTypes = {
  currentValue: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
  desiredValue: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
}

WarningTooltip.defaultProps = {
  currentValue: undefined,
  desiredValue: undefined,
}

export default WarningTooltip
