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

import React, { useCallback } from 'react'
import classnames from 'classnames'

import Overlay from '@ttn-lw/components/overlay'
import Pagination from '@ttn-lw/components/pagination'

import PropTypes from '@ttn-lw/lib/prop-types'
import getByPath from '@ttn-lw/lib/get-by-path'

import Table from './table'

import style from './tabular.styl'

const Tabular = ({
  onPageChange,
  order,
  orderBy,
  pageSize,
  setPageSize,
  page,
  handlesPagination,
  paginated,
  className,
  headerClassName,
  loading,
  small,
  onRowClick,
  totalCount,
  data,
  headers,
  rowKeySelector,
  rowHrefSelector,
  emptyMessage,
  clickable,
  disableSorting,
  onSortRequest,
  panelStyle,
}) => {
  const handlePageChange = useCallback(
    page => {
      onPageChange(page)
    },
    [onPageChange],
  )

  const handleSortRequest = useCallback(
    newOrderBy => {
      const sameColumn = orderBy === newOrderBy

      if (sameColumn && order === 'asc') {
        onSortRequest('desc', orderBy)

        return
      }

      onSortRequest('asc', newOrderBy)
    },
    [orderBy, order, onSortRequest],
  )

  const handlePagination = useCallback(
    items => {
      if (paginated && handlesPagination) {
        const from = pageSize * (page - 1)
        const to = pageSize * page

        return items.slice(from, to)
      }

      return items
    },
    [handlesPagination, page, paginated, pageSize],
  )

  const columns = (
    <Table.Row head panelStyle={panelStyle}>
      {headers.map((header, key) => (
        <Table.HeadCell
          key={key}
          align={header.align}
          content={header.sortable && !disableSorting ? undefined : header.displayName}
          name={header.name}
          width={header.width}
          className={header.className}
          panelStyle={panelStyle}
        >
          {header.sortable && !disableSorting ? (
            <Table.SortButton
              title={header.displayName}
              direction={order}
              name={
                typeof header.sortKey === 'function'
                  ? header.sortKey(header)
                  : header.sortKey || header.name
              }
              active={header.sortKey ? orderBy === header.sortKey : orderBy === header.name}
              onSort={handleSortRequest}
              align={header.align}
            />
          ) : null}
        </Table.HeadCell>
      ))}
    </Table.Row>
  )

  const minWidth = `${headers.length * 10}rem`
  const defaultRowKeySelector = row => {
    const key = headers[0].getValue ? headers[0].getValue(row) : getByPath(row, headers[0].name)
    return typeof key === 'string' || typeof key === 'number' ? key : JSON.stringify(key)
  }
  const appliedRowKeySelector = rowKeySelector ? rowKeySelector : defaultRowKeySelector
  const paginatedData = handlePagination(data)
  const rows = paginatedData.map((row, rowIndex) => {
    // If the whole table is disabled each row should be as well.
    const rowClickable = !clickable ? false : (row._meta?.clickable ?? clickable)

    return (
      <Table.Row
        key={appliedRowKeySelector(row)}
        id={rowIndex}
        onClick={onRowClick}
        clickable={rowClickable}
        linkTo={rowHrefSelector ? rowHrefSelector(row) : undefined}
        body
        panelStyle={panelStyle}
      >
        {headers.map((header, index) => {
          const value = headers[index].getValue
            ? headers[index].getValue(row)
            : getByPath(row, headers[index].name)
          return (
            <Table.DataCell key={index} align={header.align} small={small} panelStyle={panelStyle}>
              {headers[index].render ? headers[index].render(value, row) : value}
            </Table.DataCell>
          )
        })}
      </Table.Row>
    )
  })

  const pagination = paginated ? (
    <div className="d-flex justify-center">
      <Pagination
        className={style.pagination}
        pageCount={Math.ceil(totalCount / pageSize) || 1}
        onPageChange={handlePageChange}
        disableInitialCallback
        pageRangeDisplayed={2}
        forcePage={page}
        pageSize={pageSize}
        setPageSize={setPageSize}
        totalCount={totalCount}
      />
    </div>
  ) : null

  return (
    <div className={classnames(style.container, className)}>
      <Overlay visible={loading} loading={loading} className={style.overlay}>
        <Table minWidth={minWidth}>
          <Table.Head className={headerClassName} panelStyle={panelStyle}>
            {columns}
          </Table.Head>
          <Table.Body loading={loading} empty={rows.length === 0} emptyMessage={emptyMessage}>
            {rows}
          </Table.Body>
        </Table>
        <Table.Footer loading={loading} empty={rows.length === 0}>
          {pagination}
        </Table.Footer>
      </Overlay>
    </div>
  )
}

Tabular.propTypes = {
  className: PropTypes.string,
  clickable: PropTypes.bool,
  /** A list of data entries to display within the table body. */
  data: PropTypes.arrayOf(
    PropTypes.shape({
      /** A meta config object used to control the behavior of individual rows. */
      _meta: PropTypes.shape({
        /** A flag specifying whether the row should be clickable. */
        clickable: PropTypes.bool,
      }),
    }),
  ),
  /** A flag to disable any sorting in the table altogether. */
  disableSorting: PropTypes.bool,
  /** The empty message to be displayed when no data provided. */
  emptyMessage: PropTypes.message.isRequired,
  /**
   * A flag specifying whether the table should paginate entries.
   * If true the component makes sure that the items are paginated, otherwise
   * the user is responsible for passing the right number of items.
   */
  handlesPagination: PropTypes.bool,
  /** A class name to apply to the header row. */
  headerClassName: PropTypes.string,
  /** A list of head entries to display within the table head. */
  headers: PropTypes.arrayOf(
    PropTypes.shape({
      align: PropTypes.oneOf(['left', 'right', 'center']),
      displayName: PropTypes.message.isRequired,
      getValue: PropTypes.func,
      name: PropTypes.string,
      render: PropTypes.func,
      sortable: PropTypes.bool,
      sortKey: PropTypes.oneOfType([PropTypes.string, PropTypes.func]),
      width: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
      className: PropTypes.string,
    }),
  ).isRequired,
  /** A flag specifying whether the table should covered with the loading overlay. */
  loading: PropTypes.bool,
  /**
   * Function to be called when the page is changed. Passes the new
   * page number as an argument [1...pageCount - 1].
   */
  onPageChange: PropTypes.func,
  /** Function to be called when the table row gets clicked. */
  onRowClick: PropTypes.func,
  /**
   * Function to be called when the table should be sorted. Passes
   * the new ordering type and the name of the head cell that the
   * table should sorted according to.
   */
  onSortRequest: PropTypes.func,
  /** The current order of the table. */
  order: PropTypes.string,
  /** The name of the column that the table is sorted according to. */
  orderBy: PropTypes.string,
  /** The current page of the pagination. */
  page: PropTypes.number,
  /** The number of entries to display per page. */
  pageSize: PropTypes.number,
  /** A flag identifying whether the table should have pagination. */
  paginated: PropTypes.bool,
  panelStyle: PropTypes.bool,
  /** A selector to determine the `href`/`to` prop of the rendered rows. */
  rowHrefSelector: PropTypes.func,
  /** A selector to determine the `key` prop of the rendered rows. */
  rowKeySelector: PropTypes.func,
  /** A flag specifying the height of data cells. */
  setPageSize: PropTypes.func,
  small: PropTypes.bool,
  /** The total number of available entries. */
  totalCount: PropTypes.number,
}

Tabular.defaultProps = {
  className: undefined,
  data: [],
  handlesPagination: false,
  loading: false,
  onRowClick: () => null,
  onPageChange: () => null,
  onSortRequest: () => null,
  small: false,
  order: undefined,
  orderBy: undefined,
  paginated: false,
  totalCount: 0,
  page: 0,
  pageSize: undefined,
  setPageSize: () => null,
  clickable: true,
  rowKeySelector: undefined,
  rowHrefSelector: undefined,
  disableSorting: false,
  headerClassName: undefined,
  panelStyle: false,
}

export { Tabular as default, Table }
