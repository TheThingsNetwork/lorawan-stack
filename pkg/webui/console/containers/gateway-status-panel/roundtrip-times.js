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

import React, { useMemo } from 'react'
import { FormattedNumber } from 'react-intl'
import classNames from 'classnames'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './gateway-status-panel.styl'

const RoundtripTimes = ({ maxRoundTripTime, minRoundTripTime, medianRoundTripTime }) => {
  const barWidth = maxRoundTripTime - minRoundTripTime
  const position = useMemo(() => {
    const mediumPoint = medianRoundTripTime - minRoundTripTime
    return isNaN((mediumPoint * 100) / barWidth) ? 0 : (mediumPoint * 100) / barWidth
  }, [medianRoundTripTime, minRoundTripTime, barWidth])

  const greenPointer = medianRoundTripTime * 1000 <= 100
  const yellowPointer = medianRoundTripTime * 1000 > 100 && medianRoundTripTime * 1000 <= 500
  const redPointer = medianRoundTripTime * 1000 > 500

  const maxRoundTripTimeInMs = maxRoundTripTime * 1000
  const minRoundTripTimeInMs = minRoundTripTime * 1000
  const barWidthInMs = barWidth * 1000
  const greenPercentage =
    maxRoundTripTimeInMs <= 100
      ? 100
      : minRoundTripTimeInMs > 100
        ? 0
        : ((100 - minRoundTripTimeInMs) * 100) / barWidthInMs
  const yellowPercentage =
    minRoundTripTimeInMs > 100 && maxRoundTripTimeInMs < 500
      ? 100
      : minRoundTripTimeInMs > 500
        ? 0
        : maxRoundTripTimeInMs >= 500
          ? (400 * 100) / barWidthInMs
          : ((maxRoundTripTimeInMs - 100) * 100) / barWidthInMs
  const hasRed = maxRoundTripTimeInMs >= 500

  return (
    <>
      <div
        className={style.gtwStatusPanelRoundTripTimeBar}
        style={{
          background: hasRed
            ? `linear-gradient(90deg, #B3FAA9 ${greenPercentage}%, #FFE096 ${greenPercentage}%, #FFE096 ${yellowPercentage}%, #FFD1CA ${yellowPercentage}%, #FFD1CA 100%)`
            : `linear-gradient(90deg, #B3FAA9 ${greenPercentage}%, #FFE096 ${greenPercentage}%, #FFE096 ${yellowPercentage}%)`,
        }}
      >
        <span
          className={classNames(style.gtwStatusPanelRoundTripTimeBarPointer, {
            'c-bg-success-normal': greenPointer,
            'c-bg-warning-normal': yellowPointer,
            'c-bg-error-normal': redPointer,
          })}
          style={{
            left: `${position}%`,
          }}
        />
      </div>
      <div className="d-flex j-between">
        <span className="fs-s fw-bold">
          <FormattedNumber value={minRoundTripTimeInMs.toFixed(2)} />
        </span>
        <span className="fs-s fw-bold">
          <FormattedNumber value={maxRoundTripTimeInMs.toFixed(2)} />
        </span>
      </div>
      <div
        className={classNames(style.gtwStatusPanelRoundTripTimeTag, {
          'c-text-success-normal': greenPointer,
          'c-text-warning-normal': yellowPointer,
          'c-text-error-normal': redPointer,
        })}
      >
        <FormattedNumber value={(medianRoundTripTime * 1000).toFixed(2)} />
        <Message content="ms" />
      </div>
    </>
  )
}

RoundtripTimes.propTypes = {
  maxRoundTripTime: PropTypes.number.isRequired,
  medianRoundTripTime: PropTypes.number.isRequired,
  minRoundTripTime: PropTypes.number.isRequired,
}

export default RoundtripTimes
