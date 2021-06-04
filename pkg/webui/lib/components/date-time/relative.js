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
import { FormattedRelativeTime } from 'react-intl'

import PropTypes from '@ttn-lw/lib/prop-types'

import DateTime from '.'

const formatInSeconds = (from, to) => Math.floor((from - to) / 1000)

const RelativeTime = props => {
  const {
    className,
    value,
    unit,
    computeDelta,
    updateIntervalInSeconds,
    firstToLower,
    children,
  } = props

  return (
    <DateTime className={className} value={value} firstToLower={firstToLower}>
      {dateTime => {
        const from = new Date(dateTime)
        const to = new Date()

        const delta = computeDelta(from, to)

        return (
          <FormattedRelativeTime
            key={dateTime}
            value={delta}
            numeric="auto"
            updateIntervalInSeconds={updateIntervalInSeconds}
            unit={unit}
          >
            {formattedRelativeTime => children(formattedRelativeTime)}
          </FormattedRelativeTime>
        )
      }}
    </DateTime>
  )
}

RelativeTime.propTypes = {
  children: PropTypes.func,
  className: PropTypes.string,
  /** A function to compute relative delta in specified time units in the `unit` prop. */
  computeDelta: PropTypes.func,
  /** Whether to convert the first character of the resulting message to lowercase. */
  firstToLower: PropTypes.bool,
  /** The unit to calculate relative date time. */
  unit: PropTypes.oneOf(['second', 'minute', 'hour', 'day', 'week', 'month', 'year']),
  /** The interval that the component will re-render in seconds. */
  updateIntervalInSeconds: PropTypes.number,
  /** The time to be displayed. */
  value: PropTypes.oneOfType([
    PropTypes.string,
    PropTypes.number, // Support timestamps.
    PropTypes.instanceOf(Date),
  ]).isRequired,
}

RelativeTime.defaultProps = {
  children: dateTime => dateTime,
  className: undefined,
  firstToLower: true,
  updateIntervalInSeconds: 1,
  unit: 'second',
  computeDelta: formatInSeconds,
}

export default RelativeTime
