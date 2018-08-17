// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
import PropTypes from '../../lib/prop-types'

import Pagination from '../pagination'
import Message from '../../lib/components/message'
import DataCell from './data-cell'
import HeaderCell from './header-cell'

import orders from './orders'
import style from './table.styl'

const Table = function ({
  className,
  pageCount = 1,
  headers,
  rows,
  emptyMessage,
  page = 0,
  order = orders.DEFAULT,
  small = false,
  orderedBy,
  onPageChange,
  onSortByColumn,
}) {

  return (
    <div className={className}>
      <table className={style.table}>
        <thead>
          <tr className={style.headerRow}>
            {headers.map((header, index) => (
              <HeaderCell
                key={index}
                centered={header.centered}
                active={orderedBy === header.name}
                sortable={header.sortable}
                name={header.name}
                content={header.displayName}
                order={order}
                onSort={onSortByColumn}
              />
            )
            )}
          </tr>
        </thead>
        <tbody>
          {
            !!rows.length
              ? rows.map((row, index) => (
                <tr className={style.dataRow} key={index} data-hook="data-row">
                  {headers.map(function (header, idx) {
                    return (
                      <DataCell
                        key={idx}
                        centered={header.centered}
                        small={small}
                      >
                        {row[header.name]}
                      </DataCell>
                    )
                  })}
                </tr>
              ))
              : <tr>
                <td colSpan={headers.length} data-hook="empty-message">
                  <Message
                    className={style.emptyMessage}
                    content={emptyMessage}
                  />
                </td>
              </tr>
          }
        </tbody>
      </table>
      <Pagination
        className={style.pagination}
        disableInitialCallback
        pageCount={pageCount || 1}
        onPageChange={function (page) {
          onPageChange(page.selected)
        }}
        forcePage={page}
      />
    </div>
  )
}

Table.propTypes = {
  /** The current page */
  page: PropTypes.number,
  /** The total number of pages */
  pageCount: PropTypes.number,
  /** The name of a header cell according which the table is sorted */
  orderedBy: PropTypes.string,
  /** A flag specifying the height of data cells */
  small: PropTypes.bool,
  /**  The current order of the table */
  order: PropTypes.oneOf(Object.values(orders)),
  /** The header cells of the table */
  headers: PropTypes.array.isRequired,
  /** The data rows of the table */
  rows: PropTypes.array.isRequired,
  /** The empty message to be displayed when no data provided */
  emptyMessage: PropTypes.message.isRequired,
  /**
   * Function to be called when the page is changed. Passes the new
   * page number as an argument [0...pageCount - 1].
   */
  onPageChange: PropTypes.func.isRequired,
  /**
   * Function to be called when a sortable header cell is pressed.
   * Passes the name of a header cell.
   */
  onSortByColumn: PropTypes.func.isRequired,
}

export default Table
