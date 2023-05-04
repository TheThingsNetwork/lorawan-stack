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

import React, { useEffect, useState } from 'react'
import { FormattedRelativeTime, defineMessages } from 'react-intl'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import DateTime from '.'

const m = defineMessages({
  justNow: 'just now',
})

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
    justNowMessage,
    relativeTimeStyle,
    noTitle,
    showAbsoluteAfter,
    dateTimeProps,
  } = props

  const from = new Date(value)
  const to = new Date()
  const delta = computeDelta(from, to)
  const absDelta = Math.abs(delta)

  const [showLessThan, setShowLessThan] = useState(false)

  useEffect(() => {
    // Do not show the `just now` message less than 5 seconds.
    const minInterval = Math.max(updateIntervalInSeconds, 5)

    // Show the `just now` message for deltas shorter than the update interval.
    if (absDelta < minInterval) {
      setShowLessThan(true)
      const timer = setTimeout(() => {
        setShowLessThan(false)
      }, (minInterval - absDelta) * 1000)
      return () => {
        clearTimeout(timer)
      }
    }

    setShowLessThan(false)
  }, [showLessThan, absDelta, updateIntervalInSeconds])

  return (
    <DateTime
      className={className}
      value={value}
      firstToLower={firstToLower}
      noTitle={noTitle}
      {...dateTimeProps}
    >
      {dateTime =>
        absDelta >= 60 * 60 * 24 * showAbsoluteAfter ? (
          dateTime
        ) : (
          <FormattedRelativeTime
            key={dateTime}
            value={delta}
            numeric="auto"
            updateIntervalInSeconds={updateIntervalInSeconds}
            unit={unit}
            style={relativeTimeStyle}
          >
            {formattedRelativeTime =>
              showLessThan ? (
                <Message content={justNowMessage} values={{ count: updateIntervalInSeconds }} />
              ) : (
                children(formattedRelativeTime)
              )
            }
          </FormattedRelativeTime>
        )
      }
    </DateTime>
  )
}

RelativeTime.propTypes = {
  children: PropTypes.func,
  className: PropTypes.string,
  /** A function to compute relative delta in specified time units in the `unit` prop. */
  computeDelta: PropTypes.func,
  /** Passed `<DateTime />`-props for when the absolute date is rendered. */
  dateTimeProps: PropTypes.shape({}),
  /** Whether to convert the first character of the resulting message to lowercase. */
  firstToLower: PropTypes.bool,
  /** Message to render when the delta is less than `updateIntervalInSeconds`. */
  justNowMessage: PropTypes.message,
  /** Whether to show the title or not. */
  noTitle: PropTypes.bool,
  /** The style of the relative time rendering. */
  relativeTimeStyle: PropTypes.oneOf(['long', 'short', 'narrow']),
  /** After how many days the absolute date value is shown instead. */
  showAbsoluteAfter: PropTypes.number,
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
  dateTimeProps: { time: false },
  firstToLower: true,
  updateIntervalInSeconds: 1,
  relativeTimeStyle: 'long',
  unit: 'second',
  computeDelta: formatInSeconds,
  justNowMessage: m.justNow,
  noTitle: false,
  showAbsoluteAfter: 30,
}

export default RelativeTime
