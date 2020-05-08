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

import EventsHeader from '@ttn-lw/components/events-list/header'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './columns.styl'

const Column = React.memo(props => {
  const { className, ...rest } = props

  return <EventsHeader.Column className={classnames(className, style.column)} {...rest} />
})

Column.propTypes = {
  className: PropTypes.string,
}

Column.defaultProps = {
  className: undefined,
}

const IdColumn = ({ className }) => {
  return <Column className={classnames(className, style.id)} content={sharedMessages.entityId} />
}

IdColumn.propTypes = {
  className: PropTypes.string,
}

IdColumn.defaultProps = {
  className: undefined,
}

const TimeColumn = ({ className }) => {
  return <Column className={classnames(className, style.time)} content={sharedMessages.time} />
}

TimeColumn.propTypes = {
  className: PropTypes.string,
}

TimeColumn.defaultProps = {
  className: undefined,
}

const TypeColumn = ({ className }) => {
  return <Column className={classnames(className, style.type)} content={sharedMessages.type} />
}

TypeColumn.propTypes = {
  className: PropTypes.string,
}

TypeColumn.defaultProps = {
  className: undefined,
}

const DataColumn = ({ className }) => {
  return <Column className={classnames(className, style.data)} content={sharedMessages.data} />
}

DataColumn.propTypes = {
  className: PropTypes.string,
}

DataColumn.defaultProps = {
  className: undefined,
}

Column.ID = IdColumn
Column.Time = TimeColumn
Column.Type = TypeColumn
Column.Data = DataColumn

export default Column
