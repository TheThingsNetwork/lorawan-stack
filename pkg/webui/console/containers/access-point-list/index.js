// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useEffect, useState } from 'react'
import classnames from 'classnames'
import { defineMessages } from 'react-intl'

import Icon from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'

import DateTime from '@ttn-lw/lib/components/date-time'
import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './access-point-list.styl'

PropTypes.accessPoint = PropTypes.shape({
  _type: PropTypes.oneOf(['all', 'other']),
  ssid: PropTypes.string,
  password: PropTypes.string,
  security: PropTypes.string,
  signal_strength: PropTypes.number,
  is_active: PropTypes.bool,
})

const m = defineMessages({
  lastRefresh: 'Last refresh',
  description:
    'This list shows WiFi networks as detected by your gateway. You can select an access point or choose "Other..." to enter an SSID of a hidden access point.',
})

const computeDeltaInSeconds = (from, to) => {
  // Avoid situations when server clock is ahead of the browser clock.
  if (from > to) {
    return 0
  }

  return Math.floor((from - to) / 1000)
}

const AccessPointListItem = ({ accessPoint, onClick, isActive }) => {
  const handleClick = useCallback(() => {
    onClick(accessPoint)
  }, [accessPoint, onClick])
  const isOther = accessPoint._type === 'other'

  return (
    <div
      className={classnames(style.item, 'd-flex al-center j-between', {
        [style.active]: isActive,
      })}
      onClick={handleClick}
    >
      <div className="d-flex al-center gap-cs-xs">
        {!isOther && <Icon icon="wifi" />}
        {isOther ? <Message content={sharedMessages.otherOption} /> : accessPoint.ssid}
      </div>
      {Boolean(accessPoint.password) && !isOther && <Icon icon="lock" />}
    </div>
  )
}

AccessPointListItem.propTypes = {
  accessPoint: PropTypes.accessPoint.isRequired,
  isActive: PropTypes.bool.isRequired,
  onClick: PropTypes.func.isRequired,
}

const AccessPointList = ({ onChange, value, className, inputWidth, onBlur }) => {
  // TODO: Change this with selector
  const accessPoints = [
    {
      ssid: 'exampleSSID',
      password: 'examplePassword',
      security: 'WPA2',
      signal_strength: -45,
      is_active: true,
    },
    {
      ssid: 'openNetwork',
      password: '',
      security: 'None',
      signal_strength: -30,
      is_active: true,
    },
    {
      ssid: 'exampleSSID2',
      password: 'examplePassword2',
      security: 'WPA2',
      signal_strength: -45,
      is_active: true,
    },
  ]

  const [lastRefresh, setLastRefresh] = useState(new Date())

  const fetchAccessPoints = useCallback(() => {
    // TODO: Fetch access points
    setLastRefresh(new Date())
  }, [])

  useEffect(() => {
    fetchAccessPoints()
  }, [fetchAccessPoints])

  const handleSelectAccessPoint = useCallback(
    accessPoint => {
      onChange(accessPoint, true)
    },
    [onChange],
  )

  return (
    <div className={classnames(className, style.container)} onBlur={onBlur}>
      <div className="w-full">
        <div className="d-flex gap-cs-l">
          <div className={classnames(style.list, [style[`input-width-${inputWidth}`]])}>
            {accessPoints.map(a => (
              <AccessPointListItem
                key={a.ssid}
                accessPoint={{ ...a, _type: 'all' }}
                onClick={handleSelectAccessPoint}
                isActive={value.ssid === a.ssid}
              />
            ))}
            <AccessPointListItem
              accessPoint={{ ssid: '', password: '', _type: 'other' }}
              onClick={handleSelectAccessPoint}
              isActive={value._type === 'other'}
            />
          </div>
          <Message content={m.description} className="tc-subtle-gray" />
        </div>

        <div className="d-flex al-center gap-cs-s mt-cs-m">
          <Button
            type="button"
            message={sharedMessages.scanAgain}
            onClick={fetchAccessPoints}
            icon="autorenew"
          />
          <div className="tc-subtle-gray d-flex gap-cs-xxs">
            <Message content={m.lastRefresh} />
            <DateTime.Relative
              value={lastRefresh}
              computeDelta={computeDeltaInSeconds}
              relativeTimeStyle="short"
            />
          </div>
        </div>
      </div>
    </div>
  )
}

AccessPointList.propTypes = {
  className: PropTypes.string,
  inputWidth: PropTypes.inputWidth,
  onBlur: PropTypes.func.isRequired,
  onChange: PropTypes.func.isRequired,
  value: PropTypes.accessPoint.isRequired,
}

AccessPointList.defaultProps = {
  className: undefined,
  inputWidth: 'm',
}

export default AccessPointList
