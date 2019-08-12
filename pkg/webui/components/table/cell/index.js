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
import PropTypes from '../../../lib/prop-types'

import Message from '../../../lib/components/message'

import style from './cell.styl'

const Cell = function({
  className,
  component: Component,
  centered = false,
  small = false,
  colSpan = 1,
  width,
  children,
  ...rest
}) {
  const cellClassNames = classnames(className, style.cell, {
    [style.cellCentered]: centered,
    [style.cellSmall]: small,
  })

  const widthStyle = width ? { width: `${width}%` } : undefined

  return (
    <Component {...rest} style={widthStyle} className={cellClassNames} colSpan={colSpan}>
      {children}
    </Component>
  )
}

Cell.propTypes = {
  /** The width of the cell in percentages */
  width: PropTypes.number,
  /** The html name of the wrapping component */
  component: PropTypes.string.isRequired,
  /** The number of columns that the cell should occupy */
  colSpan: PropTypes.number,
  /** A flag indicating whether the data in the cell should be centered */
  centered: PropTypes.bool,
  /** A flag indicating whether the row take less height */
  small: PropTypes.bool,
}

const HeadCell = ({ className, content, children, ...rest }) => (
  <Cell className={classnames(className, style.cellHead)} component="th" {...rest}>
    {Boolean(content) && <Message content={content} />}
    {!Boolean(content) && children}
  </Cell>
)

HeadCell.propTypes = {
  /** The title of the head cell */
  content: PropTypes.message,
}

const DataCell = ({ className, children, ...rest }) => (
  <Cell className={classnames(className, style.cellData)} component="td" {...rest}>
    {children}
  </Cell>
)

export { Cell, HeadCell, DataCell }
