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
import { FormattedDate, FormattedTime } from 'react-intl'
import bind from 'autobind-decorator'

import PropTypes from '../../prop-types'
import RelativeTime from './relative'

@bind
class DateTime extends React.PureComponent {
  renderDateTime(formattedDate, formattedTime, dateValue) {
    const { className, children, date, time } = this.props

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
      <time className={className} dateTime={dateValue.toISOString()} title={result}>
        {children ? children(result) : result}
      </time>
    )
  }

  render() {
    const { value, dateFormatOptions, timeFormatOptions, dateFormat, timeFormat } = this.props

    let dateValue = value
    if (!(value instanceof Date)) {
      dateValue = new Date(value)
    }

    return (
      <FormattedDate value={dateValue} format={dateFormat} {...dateFormatOptions}>
        {date => (
          <FormattedTime value={dateValue} format={timeFormat} {...timeFormatOptions}>
            {time => this.renderDateTime(date, time, dateValue)}
          </FormattedTime>
        )}
      </FormattedDate>
    )
  }
}

DateTime.Relative = RelativeTime

DateTime.propTypes = {
  /** The time to be displayed */
  value: PropTypes.oneOfType([
    PropTypes.string,
    PropTypes.number, // support timestamps
    PropTypes.instanceOf(Date),
  ]).isRequired,
  // see https://github.com/yahoo/react-intl/wiki/Components#date-formatting-components
  dateFormatOptions: PropTypes.object,
  timeFormatOptions: PropTypes.object,
  dateFormat: PropTypes.string,
  timeFormat: PropTypes.string,
  /** Whether to show the date */
  date: PropTypes.bool,
  /** Whether to show the time */
  time: PropTypes.bool,
}

DateTime.defaultProps = {
  date: true,
  time: true,
  dateFormatOptions: {},
  timeFormatOptions: {
    hour: 'numeric',
    minute: 'numeric',
    second: 'numeric',
    hour12: false,
  },
}

export default DateTime
