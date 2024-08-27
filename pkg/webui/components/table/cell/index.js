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

import style from './cell.styl'

const Cell = ({ className, align, small, width, children, panelStyle, ...rest }) => {
  const cellClassNames = classnames(className, style.cell, {
    [style.cellPanelStyle]: panelStyle,
    [style.cellCentered]: align === 'center',
    [style.cellLeft]: align === 'left',
    [style.cellRight]: align === 'right',
    [style.cellSmall]: small,
  })

  const widthStyle =
    typeof width === 'number' ? { width: `${width}%` } : typeof width === 'string' ? { width } : {}

  return (
    <div {...rest} style={widthStyle} className={cellClassNames} role="cell">
      {children}
    </div>
  )
}

Cell.propTypes = {
  /** A flag indicating how the text in the cell should be aligned. */
  align: PropTypes.oneOf(['left', 'right', 'center']),
  children: PropTypes.node,
  className: PropTypes.string,
  /** The number of columns that the cell should occupy. */
  colSpan: PropTypes.number,
  panelStyle: PropTypes.bool,
  /** A flag indicating whether the row take less height. */
  small: PropTypes.bool,
  /** The width of the cell in percentages. */
  width: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
}

Cell.defaultProps = {
  align: undefined,
  children: undefined,
  className: undefined,
  colSpan: 1,
  small: false,
  width: undefined,
  panelStyle: false,
}

const HeadCell = ({ className, content, children, panelStyle, ...rest }) => (
  <Cell
    className={classnames(className, style.cellHead, { [style.cellHeadPanelStyle]: panelStyle })}
    panelStyle={panelStyle}
    {...rest}
  >
    {Boolean(content) && <Message content={content} />}
    {!Boolean(content) && children}
  </Cell>
)

HeadCell.propTypes = {
  children: PropTypes.node,
  className: PropTypes.string,
  /** The title of the head cell. */
  content: PropTypes.message,
  /** A flag indicating whether the table is panel styled. */
  panelStyle: PropTypes.bool,
}

HeadCell.defaultProps = {
  children: undefined,
  className: undefined,
  content: undefined,
  panelStyle: false,
}

const DataCell = ({ className, children, ...rest }) => (
  <Cell className={classnames(className, style.cellData)} {...rest}>
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
