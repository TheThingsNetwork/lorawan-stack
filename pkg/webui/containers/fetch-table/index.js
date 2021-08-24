// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import React, { Component } from 'react'
import { defineMessages } from 'react-intl'
import { connect } from 'react-redux'
import { push } from 'connected-react-router'
import bind from 'autobind-decorator'
import classnames from 'classnames'

import PAGE_SIZES from '@ttn-lw/constants/page-sizes'

import Tabular from '@ttn-lw/components/table'
import Input from '@ttn-lw/components/input'
import Button from '@ttn-lw/components/button'
import Tabs from '@ttn-lw/components/tabs'
import Overlay from '@ttn-lw/components/overlay'
import ErrorNotification from '@ttn-lw/components/error-notification'

import debounce from '@ttn-lw/lib/debounce'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import style from './fetch-table.styl'

const DEFAULT_PAGE = 1

const filterValidator = filters => {
  if (typeof filters.order === 'string' && filters.order.match(/-?[a-z0-9]/) === null) {
    filters.order = undefined
  }

  if (!Boolean(filters.page) || filters.page < 0) {
    filters.page = DEFAULT_PAGE
  }

  return filters
}

const m = defineMessages({
  errorMessage: `There was an error and the list of {entity, select,
    applications {applications}
    organizations {organizations}
    keys {API keys}
    collaborators {collaborators}
    devices {end devices}
    gateways {gateways}
    users {users}
    webhooks {webhooks}
    other {entities}
  } could not be displayed`,
})

@connect((state, props) => {
  const base = props.baseDataSelector(state, props)

  return {
    items: base[props.entity] || [],
    totalCount: base.totalCount || 0,
    fetching: base.fetching,
    fetchingSearch: base.fetchingSearch,
    pathname: state.router.location.pathname,
    mayAdd: 'mayAdd' in base ? base.mayAdd : true,
    error: base.error,
  }
})
class FetchTable extends Component {
  static propTypes = {
    actionItems: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]),
    addMessage: PropTypes.message,
    clickable: PropTypes.bool,
    dispatch: PropTypes.func.isRequired,
    entity: PropTypes.string.isRequired,
    error: PropTypes.error,
    fetching: PropTypes.bool,
    fetchingSearch: PropTypes.bool,
    filterValidator: PropTypes.func,
    getItemPathPrefix: PropTypes.func,
    getItemsAction: PropTypes.func.isRequired,
    handlesPagination: PropTypes.bool,
    headers: PropTypes.arrayOf(
      PropTypes.shape({
        displayName: PropTypes.message.isRequired,
        getValue: PropTypes.func,
        name: PropTypes.string,
        render: PropTypes.func,
        align: PropTypes.oneOf(['left', 'right', 'center']),
        sortable: PropTypes.bool,
        width: PropTypes.number,
      }),
    ),
    itemPathPrefix: PropTypes.string,
    items: PropTypes.arrayOf(
      PropTypes.shape({
        id: PropTypes.oneOfType([PropTypes.string, PropTypes.shape({})]),
        ids: PropTypes.shape({}),
      }),
    ),
    mayAdd: PropTypes.bool,
    pageSize: PropTypes.number,
    pathname: PropTypes.string.isRequired,
    searchItemsAction: PropTypes.func,
    searchPlaceholderMessage: PropTypes.message,
    searchable: PropTypes.bool,
    tableTitle: PropTypes.message,
    tabs: PropTypes.arrayOf(
      PropTypes.shape({
        title: PropTypes.message.isRequired,
        name: PropTypes.string.isRequired,
        icon: PropTypes.string,
        disabled: PropTypes.bool,
      }),
    ),
    totalCount: PropTypes.number,
  }

  static defaultProps = {
    getItemPathPrefix: undefined,
    searchItemsAction: undefined,
    pageSize: PAGE_SIZES.REGULAR,
    filterValidator,
    itemPathPrefix: '',
    mayAdd: false,
    searchable: false,
    searchPlaceholderMessage: sharedMessages.searchById,
    handlesPagination: false,
    fetching: false,
    totalCount: 0,
    items: [],
    headers: [],
    fetchingSearch: false,
    addMessage: undefined,
    tableTitle: undefined,
    tabs: [],
    actionItems: null,
    error: undefined,
    clickable: true,
  }

  constructor(props) {
    super(props)

    const { tabs } = props

    this.state = {
      query: '',
      page: 1,
      tab: tabs.length > 0 ? tabs[0].name : undefined,
      order: undefined,
      initialFetch: true,
    }

    const { debounced: debouncedFunction, cancel: cancelFunction } = debounce(
      this.requestSearch,
      350,
    )

    this.debouncedRequestSearch = debouncedFunction
    this.debounceCancel = cancelFunction
  }

  async componentDidMount() {
    await this.fetchItems(true)
    this.setState({ initialFetch: false })
  }

  componentWillUnmount() {
    this.debounceCancel()
  }

  @bind
  fetchItems() {
    const { dispatch, pageSize, searchItemsAction, getItemsAction } = this.props

    const filters = { ...this.state, limit: pageSize }

    if (filters.query && searchItemsAction) {
      return dispatch(attachPromise(searchItemsAction(filters)))
    }

    return dispatch(attachPromise(getItemsAction(filters)))
  }

  @bind
  async onPageChange(page) {
    await this.setState(
      this.props.filterValidator({
        ...this.state,
        page,
      }),
    )

    this.fetchItems()
  }

  @bind
  async requestSearch() {
    await this.setState(
      this.props.filterValidator({
        ...this.state,
        page: 1,
      }),
    )

    this.fetchItems()
  }

  @bind
  async onQueryChange(query) {
    await this.setState(
      this.props.filterValidator({
        ...this.state,
        query,
      }),
    )

    this.debouncedRequestSearch()
  }

  @bind
  async onOrderChange(order, orderBy) {
    const filterOrder = `${order === 'desc' ? '-' : ''}${orderBy}`

    await this.setState(
      this.props.filterValidator({
        ...this.state,
        order: filterOrder,
      }),
    )

    this.fetchItems()
  }

  @bind
  async onTabChange(tab) {
    await this.setState(
      this.props.filterValidator({
        ...this.state,
        query: '',
        page: 1,
        tab,
      }),
    )
    this.fetchItems()
  }

  @bind
  onItemClick(index) {
    const {
      dispatch,
      pathname,
      items,
      entity,
      itemPathPrefix,
      getItemPathPrefix,
      handlesPagination,
      pageSize,
    } = this.props
    const { page } = this.state

    let itemIndex = index
    if (handlesPagination) {
      const pageNr = page - 1 // Switch to 0-based pagination.
      itemIndex += pageSize * pageNr
    }

    const entitySingle = entity.substr(0, entity.length - 1)
    let entityPath
    if (Boolean(getItemPathPrefix)) {
      entityPath = getItemPathPrefix(items[itemIndex])
    } else {
      const item_id = items[itemIndex].id || items[itemIndex].ids[`${entitySingle}_id`]
      entityPath = `${itemPathPrefix}/${item_id}`
    }

    dispatch(push(`${pathname}${entityPath}`))
  }

  render() {
    const {
      items,
      totalCount,
      fetching,
      fetchingSearch,
      mayAdd,
      pageSize,
      addMessage,
      tableTitle,
      headers,
      tabs,
      searchable,
      handlesPagination,
      itemPathPrefix,
      pathname,
      actionItems,
      entity,
      error,
      searchPlaceholderMessage,
      clickable,
    } = this.props
    const { page, query, tab, order, initialFetch } = this.state
    let orderDirection, orderBy

    // Parse order string.
    if (typeof order === 'string') {
      orderDirection = typeof order === 'string' && order[0] === '-' ? 'desc' : 'asc'
      orderBy = typeof order === 'string' && order[0] === '-' ? order.substr(1) : order
    }

    const filtersCls = classnames(style.filters, {
      [style.topRule]: tabs.length > 0,
    })

    return (
      <div data-test-id={`${entity}-table`}>
        <div className={filtersCls}>
          <div className={style.filtersLeft}>
            {tabs.length > 0 ? (
              <Tabs
                active={tab}
                className={style.tabs}
                tabs={tabs}
                onTabChange={this.onTabChange}
              />
            ) : (
              tableTitle && (
                <div className={style.tableTitle}>
                  {tableTitle} ({totalCount})
                </div>
              )
            )}
          </div>
          <div className={style.filtersRight}>
            {searchable && (
              <Input
                data-test-id="search-input"
                value={query}
                icon="search"
                loading={fetchingSearch}
                onChange={this.onQueryChange}
                placeholder={searchPlaceholderMessage}
                className={style.searchBar}
                inputWidth="full"
              />
            )}
            {(Boolean(actionItems) || mayAdd) && (
              <div className={style.actionItems}>
                {actionItems}
                {mayAdd && (
                  <Button.Link
                    className={style.addButton}
                    message={addMessage}
                    icon="add"
                    to={`${pathname}${itemPathPrefix}/add`}
                  />
                )}
              </div>
            )}
          </div>
        </div>
        <Overlay visible={Boolean(error)}>
          {Boolean(error) && (
            <ErrorNotification
              className={style.errorMessage}
              content={{ ...m.errorMessage, values: { entity } }}
            />
          )}
          <Tabular
            paginated
            page={page}
            totalCount={totalCount}
            pageSize={pageSize}
            onRowClick={this.onItemClick}
            onPageChange={this.onPageChange}
            loading={fetching}
            headers={headers}
            data={initialFetch ? [] : items}
            emptyMessage={sharedMessages.noMatch}
            handlesPagination={handlesPagination}
            onSortRequest={this.onOrderChange}
            order={orderDirection}
            orderBy={orderBy}
            clickable={clickable}
          />
        </Overlay>
      </div>
    )
  }
}

export default FetchTable
