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
import { storiesOf } from '@storybook/react'

import RelativeDateTime from './relative'

const Example = ({ title, unit, ago }) => {
  const date = new Date()
  if (unit === 'hour') {
    date.setHours(date.getHours() - ago)
  } else if (unit === 'minute') {
    date.setMinutes(date.getMinutes() - ago)
  } else if (unit === 'day') {
    date.setDate(date.getDate() - ago)
  } else if (unit === 'month') {
    date.setMonth(date.getMonth() - ago)
  } else if (unit === 'year') {
    date.setFullYear(date.getFullYear() - ago)
  }

  return (
    <div>
      <h3>{title}:</h3>
      <RelativeDateTime value={date} />
      <hr />
    </div>
  )
}

storiesOf('DateTime/Relative', module).add('Default', () => (
  <div>
    <Example title="from now" />
    <Example title="from 1 minute ago" unit="minute" ago={1} />
    <Example title="from 30 minutes ago" unit="minute" ago={30} />
    <Example title="from 1 hour ago" unit="hour" ago={1} />
    <Example title="from 12 hours ago" unit="hour" ago={12} />
    <Example title="from 1 day ago" unit="day" ago={1} />
    <Example title="from 7 days ago" unit="day" ago={7} />
    <Example title="from 1 month ago" unit="month" ago={1} />
    <Example title="from 6 months ago" unit="month" ago={6} />
    <Example title="from 1 year ago" unit="year" ago={1} />
    <Example title="from 2 years ago" unit="year" ago={2} />
  </div>
))
