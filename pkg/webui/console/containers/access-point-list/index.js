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

import React, { useCallback, useEffect, useRef, useState } from 'react'
import classnames from 'classnames'
import { defineMessages } from 'react-intl'
import { useDispatch, useSelector } from 'react-redux'

import Icon from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import Spinner from '@ttn-lw/components/spinner'

import DateTime from '@ttn-lw/lib/components/date-time'
import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { selectFetchingEntry } from '@ttn-lw/lib/store/selectors/fetching'

import { GET_ACCESS_POINTS_BASE, getAccessPoints } from '@console/store/actions/connection-profiles'

import { selectSelectedGateway } from '@console/store/selectors/gateways'
import { selectAccessPoints } from '@console/store/selectors/connection-profiles'

import style from './access-point-list.styl'

PropTypes.accessPoint = PropTypes.shape({
  type: PropTypes.oneOf(['all', 'other']),
  ssid: PropTypes.string,
  bssid: PropTypes.string,
  channel: PropTypes.number,
  authentication_mode: PropTypes.string,
  rssi: PropTypes.number,
  is_password_set: PropTypes.bool,
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

const wifiIconBasedOnRSSI = rssi => {
  if (rssi >= -60) {
    // Strong
    return 'wifi'
  } else if (rssi >= -80) {
    // Moderate
    return 'signal_wifi_2'
  }
  // Weak
  return 'signal_wifi_1'
}

const AccessPointListItem = ({ accessPoint, onClick, isActive }) => {
  const handleClick = useCallback(() => {
    onClick(accessPoint)
  }, [accessPoint, onClick])
  const isOther = accessPoint.type === 'other'

  return (
    <div
      className={classnames(style.item, 'd-flex al-center j-between', {
        [style.active]: isActive,
      })}
      onClick={handleClick}
    >
      <div className="d-flex al-center gap-cs-xs">
        {!isOther && <Icon icon={wifiIconBasedOnRSSI(accessPoint.rssi)} />}
        {isOther ? <Message content={sharedMessages.otherOption} /> : accessPoint.ssid}
      </div>
      {accessPoint.authentication_mode !== 'open' && !isOther && (
        <Icon icon="lock" small className="tc-subtle-gray" />
      )}
    </div>
  )
}

AccessPointListItem.propTypes = {
  accessPoint: PropTypes.accessPoint.isRequired,
  isActive: PropTypes.bool.isRequired,
  onClick: PropTypes.func.isRequired,
}

const AccessPointList = ({ onChange, value, className, inputWidth, onBlur, ssid }) => {
  const [lastRefresh, setLastRefresh] = useState(undefined)

  const dispatch = useDispatch()
  const isLoading = useSelector(state => selectFetchingEntry(state, GET_ACCESS_POINTS_BASE))
  const accessPoints = useSelector(selectAccessPoints)
  const selectedGateway = useSelector(selectSelectedGateway)
  const { ids } = selectedGateway
  const isFirstRender = useRef(true)

  const [isMounted, setIsMounted] = useState(true)

  const handleScanAccessPoints = useCallback(() => {
    dispatch(attachPromise(getAccessPoints(ids.gateway_id, ids.eui))).then(() => {
      if (isMounted) {
        setLastRefresh(new Date())
      }
    })
  }, [dispatch, ids.eui, ids.gateway_id, isMounted])

  // Trigger this useEffect only the first time component is rendered and has accessPoints loaded.
  // Change value only if ssid is provided. (if it's edit form)
  useEffect(() => {
    if (!isFirstRender.current && ssid && !isLoading) {
      const accessPoint = accessPoints.find(ap => ap.ssid === ssid)
      const updatedAccessPoint = {
        ...accessPoint,
        is_password_set: true,
        type: accessPoint ? 'all' : 'other',
      }
      onChange(updatedAccessPoint, true)
    }
    isFirstRender.current = false
    /* eslint-disable */
  }, [accessPoints, isLoading])

  useEffect(() => {
    handleScanAccessPoints()

    return () => {
      setIsMounted(false)
    }
  }, [handleScanAccessPoints])

  const handleSelectAccessPoint = useCallback(
    accessPoint => {
      onChange(accessPoint, true)
    },
    [onChange],
  )

  return (
    <div className={classnames(className, 'd-flex', 'w-full')} onBlur={onBlur}>
      <div className="w-full">
        {isLoading ? (
          <div className="d-flex mt-cs-m">
            <Spinner>
              <Message content={sharedMessages.fetching} />
            </Spinner>
          </div>
        ) : (
          <div className="d-flex gap-cs-l">
            <div className={classnames(style.list, [style[`input-width-${inputWidth}`]])}>
              {accessPoints.map(a => (
                <AccessPointListItem
                  key={a.bssid}
                  accessPoint={{ ...a, is_password_set: false, type: 'all' }}
                  onClick={handleSelectAccessPoint}
                  isActive={value.bssid === a.bssid}
                />
              ))}
              <AccessPointListItem
                accessPoint={{ ssid: '', is_password_set: false, type: 'other' }}
                onClick={handleSelectAccessPoint}
                isActive={value.type === 'other'}
              />
            </div>
            <Message content={m.description} className="tc-subtle-gray" />
          </div>
        )}

        <div className="d-flex al-center gap-cs-s mt-cs-m">
          <Button
            type="button"
            message={sharedMessages.scanAgain}
            onClick={handleScanAccessPoints}
            icon="autorenew"
            disabled={isLoading}
          />
          {Boolean(lastRefresh) && (
            <div className="tc-subtle-gray d-flex gap-cs-xxs">
              <Message content={m.lastRefresh} />
              <DateTime.Relative
                value={lastRefresh}
                computeDelta={computeDeltaInSeconds}
                relativeTimeStyle="short"
              />
            </div>
          )}
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
  ssid: PropTypes.string,
  value: PropTypes.accessPoint.isRequired,
}

AccessPointList.defaultProps = {
  className: undefined,
  inputWidth: 'm',
  ssid: undefined,
}

export default AccessPointList
