// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import classnames from 'classnames'

import Icon from '@ttn-lw/components/icon'
import Event from '@ttn-lw/components/events-list/event'

import ErrorMessage from '@ttn-lw/lib/components/error-message'
import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import { getBackendErrorRootCause } from '@ttn-lw/lib/errors/utils'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './entries.styl'

const { Overview } = Event

const Entry = React.memo(props => {
  return <Overview.Entry {...props} />
})

const IconEntry = ({ className, iconName }) => {
  return (
    <Entry className={classnames(className, style.icon)}>
      <Icon icon={iconName} />
    </Entry>
  )
}

IconEntry.propTypes = {
  className: PropTypes.string,
  iconName: PropTypes.string,
}

IconEntry.defaultProps = {
  className: undefined,
  iconName: 'event',
}

const TimeEntry = ({ className, time }) => {
  return (
    <Entry className={classnames(className, style.time)}>
      <DateTime value={time} date={false} />
    </Entry>
  )
}

TimeEntry.propTypes = {
  className: PropTypes.string,
  time: PropTypes.string.isRequired,
}

TimeEntry.defaultProps = {
  className: undefined,
}

const IdEntry = ({ className, entityId }) => {
  return (
    <Entry className={classnames(className, style.id)}>
      <span className={style.truncate}>{entityId}</span>
    </Entry>
  )
}

IdEntry.propTypes = {
  className: PropTypes.string,
  entityId: PropTypes.string.isRequired,
}

IdEntry.defaultProps = {
  className: undefined,
}

const TypeEntry = ({ className, eventName }) => {
  return (
    <Entry className={classnames(className, style.type)}>
      <Message className={style.truncate} firstToUpper content={{ id: `event:${eventName}` }} />
    </Entry>
  )
}

TypeEntry.propTypes = {
  className: PropTypes.string,
  eventName: PropTypes.string.isRequired,
}

TypeEntry.defaultProps = {
  className: undefined,
}

const DataEntry = ({ className, children }) => {
  return <Entry className={classnames(className, style.data)}>{children}</Entry>
}

DataEntry.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]),
  className: PropTypes.string,
}

DataEntry.defaultProps = {
  children: null,
  className: undefined,
}

const ErrorDescriptionEntry = ({ className, errorDetails }) => {
  // Transform error details to error-like structure.
  const rootCause = getBackendErrorRootCause({ details: [errorDetails] })

  return (
    <Entry className={classnames(className, style.errorDescription)}>
      <ErrorMessage className={style.truncate} content={rootCause} />
    </Entry>
  )
}

ErrorDescriptionEntry.propTypes = {
  className: PropTypes.string,
  errorDetails: PropTypes.shape({
    namespace: PropTypes.string.isRequired,
    name: PropTypes.string.isRequired,
    message_format: PropTypes.string.isRequired,
    attributes: PropTypes.shape({}),
    code: PropTypes.number.isRequired,
  }).isRequired,
}

ErrorDescriptionEntry.defaultProps = {
  className: undefined,
}

Entry.Icon = IconEntry
Entry.ID = IdEntry
Entry.Time = TimeEntry
Entry.Type = TypeEntry
Entry.Data = DataEntry
Entry.ErrorDescription = ErrorDescriptionEntry

export default Entry
