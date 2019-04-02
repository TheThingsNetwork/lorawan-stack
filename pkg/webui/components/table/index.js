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
import bind from 'autobind-decorator'
import PropTypes from '../../lib/prop-types'
import getByPath from '../../lib/get-by-path'

import Overlay from '../overlay'
import Pagination from '../pagination'
import Table from './table'

import style from './tabular.styl'

@bind
class Tabular extends React.Component {

  onPageChange (page) {
    this.props.onPageChange(page)
  }

  onSortRequest (newOrderBy) {
    const { order, orderBy } = this.props
    const sameColumn = orderBy === newOrderBy

    if (sameColumn && order === 'asc') {
      this.props.onSortRequest('desc', orderBy)

      return
    } else if (sameColumn && order === 'desc') {
      this.props.onSortRequest()

      return
    }

    this.props.onSortRequest('asc', newOrderBy)
  }

  handlePagination (items = []) {
    const { pageSize, page, handlesPagination, paginated } = this.props

    if (paginated && handlesPagination) {
      const from = pageSize * (page - 1)
      const to = pageSize * page

      return items.slice(from, to)
    }

    return items
  }

  render () {
    const {
      className,
      loading = false,
      small = false,
      onRowClick,
      page,
      order = undefined,
      orderBy = undefined,
      totalCount,
      pageSize,
      initialPage = 1,
      paginated = false,
      data,
      headers,
      emptyMessage,
    } = this.props

    const columns = (
      <Table.Row>
        {
          headers.map((header, key) => (
            <Table.HeadCell
              key={key}
              centered={header.centered}
              content={header.sortable ? undefined : header.displayName}
              name={header.name}
              width={header.width}
            >
              {
                header.sortable ? (
                  <Table.SortButton
                    title={header.displayName}
                    direction={order}
                    name={header.name}
                    active={orderBy === header.name}
                    onSort={this.onSortRequest}
                  />
                ) : null
              }
            </Table.HeadCell>
          ))
        }
      </Table.Row>
    )

    const paginatedData = this.handlePagination(data)
    const rows = paginatedData.length > 0 ? (
      paginatedData.map( function (row, rowKey) {
        return (
          <Table.Row
            key={rowKey}
            id={rowKey}
            onClick={onRowClick}
          >
            {
              headers.map(function (header, index) {
                const value = getByPath(row, headers[index].name)
                return (
                  <Table.DataCell
                    key={index}
                    centered={header.centered}
                    small={small}
                  >
                    {headers[index].render
                      ? headers[index].render(value)
                      : value
                    }
                  </Table.DataCell>
                )
              })
            }
          </Table.Row>
        )
      })
    ) : (
      <Table.Empty
        colSpan={headers.length}
        message={emptyMessage}
      />
    )

    const pagination = paginated ? (
      <Table.Row>
        <Table.DataCell
          className={style.paginationCell}
          colSpan={headers.length}
          small={small}
        >
          <Pagination
            className={style.pagination}
            pageCount={Math.ceil(totalCount / pageSize) || 1}
            initialPage={initialPage}
            onPageChange={this.onPageChange}
            disableInitialCallback
            pageRangeDisplayed={2}
            forcePage={page}
          />
        </Table.DataCell>
      </Table.Row>
    ) : null

    return (
      <div className={className}>
        <Overlay visible={loading} loading={loading}>
          <Table>
            <Table.Head>
              {columns}
            </Table.Head>
            <Table.Body>
              {rows}
            </Table.Body>
            <Table.Footer>
              {pagination}
            </Table.Footer>
          </Table>
        </Overlay>
      </div>
    )
  }
}

Tabular.propTypes = {
  /** The current page of the pagination*/
  page: PropTypes.number,
  /** The initial page of pagination */
  initialPage: PropTypes.number,
  /** The total number of available entries */
  totalCount: PropTypes.number,
  /** The number of entries to display per page */
  pageSize: PropTypes.number,
  /** A flag identifying whether the table should have pagination */
  paginated: PropTypes.bool,
  /** A flag specifying the height of data cells */
  small: PropTypes.bool,
  /** The current order of the table */
  order: PropTypes.string,
  /** The name of the column that the table is sorted according to */
  orderBy: PropTypes.string,
  /** The empty message to be displayed when no data provided */
  emptyMessage: PropTypes.oneOfType([ PropTypes.message, PropTypes.string ]).isRequired,
  /** Function to be called when the table row gets clicked */
  onRowClick: PropTypes.func,
  /**
   * Function to be called when the page is changed. Passes the new
   * page number as an argument [1...pageCount - 1].
   */
  onPageChange: PropTypes.func,
  /**
   * Function to be called when the table should be sorted. Passes
   * the new ordering type and the name of the head cell that the
   * table should sorted according to.
   */
  onSortRequest: PropTypes.func,
  /** A flag specifying whether the table should covered with the loading overlay */
  loading: PropTypes.bool,
  /** A list of data entries to display within the table body */
  data: PropTypes.arrayOf(PropTypes.object),
  /** A list of head entries to displat within the table head */
  headers: PropTypes.arrayOf(PropTypes.shape({
    displayName: PropTypes.message.isRequired,
    name: PropTypes.string.isRequired,
    render: PropTypes.func,
    centered: PropTypes.bool,
    sortable: PropTypes.bool,
    width: PropTypes.number,
  })),
  /**
   * A flag specifying whether the table should paginate entries.
   * If true the component makes sure that the items are paginated, otherwise
   * the user is responsible for passing the right number of items.
   */
  handlesPagination: PropTypes.bool,
}

export { Tabular as default, Table }
