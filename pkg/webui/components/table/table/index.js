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
import classnames from 'classnames'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import SortButton from '../sort-button'
import Row from '../row'
import { HeadCell, DataCell } from '../cell'

import style from './table.styl'

/* Empty message to render when no entries provided. */
const Empty = ({ message }) => (
  <Message className={style.emptyMessage} component="div" content={message} />
)

Empty.propTypes = {
  message: PropTypes.message,
}

Empty.defaultProps = {
  message: undefined,
}

const Head = ({ className, ...props }) => (
  <div {...props} className={classnames(className, style.sectionHeader)} />
)

Head.propTypes = {
  className: PropTypes.string,
}

Head.defaultProps = {
  className: undefined,
}

const Body = ({ className, empty, emptyMessage, ...props }) => {
  if (empty) {
    return <Empty message={emptyMessage} />
  }

  return <div {...props} className={classnames(className, style.sectionBody)} role="rowgroup" />
}

Body.propTypes = {
  className: PropTypes.string,
  empty: PropTypes.bool,
  emptyMessage: PropTypes.message,
}

Body.defaultProps = {
  className: undefined,
  empty: false,
  emptyMessage: undefined,
}

const Footer = ({ className, ...props }) => (
  <div {...props} className={classnames(className, style.sectionFooter)} />
)

Footer.propTypes = {
  className: PropTypes.string,
}

Footer.defaultProps = {
  className: undefined,
}

const Table = ({ className, children, minWidth, ...rest }) => {
  const tableClassNames = classnames(className, style.table)
  const minWidthProp = Boolean(minWidth) ? { style: { minWidth } } : {}
  return (
    <div role="table" className={tableClassNames} {...minWidthProp} {...rest}>
      {children}
    </div>
  )
}

Table.propTypes = {
  children: PropTypes.node,
  className: PropTypes.string,
  minWidth: PropTypes.string,
}

Table.defaultProps = {
  className: undefined,
  children: undefined,
  minWidth: undefined,
}

Table.Row = Row
Table.HeadCell = HeadCell
Table.DataCell = DataCell
Table.SortButton = SortButton
Table.Head = Head
Table.Body = Body
Table.Footer = Footer

export default Table
