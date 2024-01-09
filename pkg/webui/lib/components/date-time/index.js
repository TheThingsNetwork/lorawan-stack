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

import React, { useCallback } from 'react'
import { FormattedDate, FormattedTime } from 'react-intl'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { warn } from '@ttn-lw/lib/log'

import RelativeTime from './relative'

const DateTime = props => {
  const {
    value,
    date,
    time,
    className,
    children,
    dateFormatOptions,
    timeFormatOptions,
    firstToLower,
    noTitle,
  } = props

  const renderUnknown = (
    <time className={className}>
      <Message content={sharedMessages.unknown} firstToLower={firstToLower} />
    </time>
  )

  const renderDateTime = useCallback(
    (formattedDate, formattedTime, dateValue) => {
      let result = ''
      if (date) {
        result += formattedDate
      }

      if (time) {
        if (date) {
          result += ' '
        }
        result += formattedTime
      }

      return (
        <time
          className={className}
          dateTime={dateValue.toISOString()}
          title={noTitle ? undefined : result}
        >
          {children ? children(result) : result}
        </time>
      )
    },
    [children, className, date, noTitle, time],
  )

  try {
    let dateValue = value
    if (!(value instanceof Date)) {
      dateValue = new Date(value)
    }

    return (
      <FormattedDate value={dateValue} {...dateFormatOptions}>
        {date => (
          <FormattedTime value={dateValue} {...timeFormatOptions}>
            {time => renderDateTime(date, time, dateValue)}
          </FormattedTime>
        )}
      </FormattedDate>
    )
  } catch (error) {
    warn(`Error rendering date time with value: "${value}"`, error)
    return renderUnknown
  }
}

DateTime.Relative = RelativeTime

DateTime.propTypes = {
  children: PropTypes.func,
  className: PropTypes.string,
  date: PropTypes.bool,
  dateFormatOptions: PropTypes.shape({}),
  firstToLower: PropTypes.bool,
  noTitle: PropTypes.bool,
  time: PropTypes.bool,
  timeFormatOptions: PropTypes.shape({}),
  value: PropTypes.oneOfType([PropTypes.string, PropTypes.number, PropTypes.instanceOf(Date)])
    .isRequired,
}

DateTime.defaultProps = {
  className: undefined,
  children: undefined,
  date: true,
  time: true,
  firstToLower: true,
  dateFormatOptions: {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  },
  timeFormatOptions: {
    hour: 'numeric',
    minute: 'numeric',
    second: 'numeric',
    hourCycle: 'h23',
  },
  noTitle: false,
}

export default DateTime
