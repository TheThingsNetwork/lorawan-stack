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
  centered,
  small,
  colSpan,
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
  /** A flag indicating whether the data in the cell should be centered */
  centered: PropTypes.bool,
  children: PropTypes.node,
  className: PropTypes.string,
  /** The number of columns that the cell should occupy */
  colSpan: PropTypes.number,
  /** The html name of the wrapping component */
  component: PropTypes.string.isRequired,
  /** A flag indicating whether the row take less height */
  small: PropTypes.bool,
  /** The width of the cell in percentages */
  width: PropTypes.number,
}

Cell.defaultProps = {
  centered: false,
  children: undefined,
  className: undefined,
  colSpan: 1,
  small: false,
  width: undefined,
}

const HeadCell = ({ className, content, children, ...rest }) => (
  <Cell className={classnames(className, style.cellHead)} component="th" {...rest}>
    {Boolean(content) && <Message content={content} />}
    {!Boolean(content) && children}
  </Cell>
)

HeadCell.propTypes = {
  children: PropTypes.node,
  className: PropTypes.string,
  /** The title of the head cell */
  content: PropTypes.message,
}

HeadCell.defaultProps = {
  children: undefined,
  className: undefined,
  content: undefined,
}

const DataCell = ({ className, children, ...rest }) => (
  <Cell className={classnames(className, style.cellData)} component="td" {...rest}>
    {children}
  </Cell>
)

DataCell.propTypes = {
  children: PropTypes.node,
  className: PropTypes.string,
}

DataCell.defaultProps = {
  children: undefined,
  className: undefined,
}

export { Cell, HeadCell, DataCell }
