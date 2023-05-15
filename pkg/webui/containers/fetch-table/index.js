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
import bind from 'autobind-decorator'
import classnames from 'classnames'
import { orderBy as lodashOrderBy } from 'lodash'

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
import getByPath from '@ttn-lw/lib/get-by-path'

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
    mayLink: 'mayLink' in base ? base.mayLink : true,
    error: base.error,
  }
})
class FetchTable extends Component {
  static propTypes = {
    actionItems: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]),
    addMessage: PropTypes.message,
    baseDataSelector: PropTypes.func.isRequired,
    clickable: PropTypes.bool,
    defaultOrder: PropTypes.string,
    dispatch: PropTypes.func.isRequired,
    entity: PropTypes.string.isRequired,
    fetching: PropTypes.bool,
    fetchingSearch: PropTypes.bool,
    filterValidator: PropTypes.func,
    getItemPathPrefix: PropTypes.func,
    getItemsAction: PropTypes.func.isRequired,
    handlesPagination: PropTypes.bool,
    handlesSorting: PropTypes.bool,
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
    mayLink: PropTypes.bool,
    pageSize: PropTypes.number,
    paginated: PropTypes.bool,
    pathname: PropTypes.string.isRequired,
    rowKeySelector: PropTypes.func,
    searchItemsAction: PropTypes.func,
    searchPlaceholderMessage: PropTypes.message,
    searchQueryMaxLength: PropTypes.number,
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
    mayLink: false,
    searchable: false,
    searchPlaceholderMessage: sharedMessages.search,
    searchQueryMaxLength: 50,
    paginated: true,
    handlesPagination: false,
    handlesSorting: false,
    fetching: false,
    totalCount: 0,
    items: [],
    headers: [],
    rowKeySelector: undefined,
    fetchingSearch: false,
    addMessage: undefined,
    tableTitle: undefined,
    tabs: [],
    actionItems: null,
    clickable: true,
    defaultOrder: undefined,
  }

  constructor(props) {
    super(props)

    const { tabs, defaultOrder } = props

    this.state = {
      query: '',
      page: 1,
      tab: tabs.length > 0 ? tabs[0].name : undefined,
      order: defaultOrder,
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
  async fetchItems() {
    const { dispatch, pageSize, searchItemsAction, getItemsAction } = this.props

    const filters = { ...this.state, limit: pageSize }

    try {
      if (filters.query && searchItemsAction) {
        await dispatch(attachPromise(searchItemsAction(filters)))
      }

      await dispatch(attachPromise(getItemsAction(filters)))
    } catch (error) {
      this.setState({ error })
    }
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
    const { handlesSorting, totalCount, pageSize } = this.props
    const canHandleSorting = totalCount <= pageSize
    const handleSorting = handlesSorting && canHandleSorting

    await this.setState(
      this.props.filterValidator({
        ...this.state,
        order: filterOrder,
      }),
    )

    if (!handleSorting) {
      this.fetchItems()
    }
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
  rowHrefSelector(item) {
    const { pathname, entity, itemPathPrefix, getItemPathPrefix } = this.props

    const entitySingle = entity.substr(0, entity.length - 1)
    let entityPath
    if (Boolean(getItemPathPrefix)) {
      entityPath = getItemPathPrefix(item)
    } else {
      const item_id = item.id || item.ids[`${entitySingle}_id`]
      entityPath = `${itemPathPrefix}/${item_id}`
    }

    return `${pathname}${entityPath}`
  }

  render() {
    const {
      items,
      totalCount,
      fetching,
      fetchingSearch,
      mayAdd,
      mayLink,
      pageSize,
      addMessage,
      tableTitle,
      headers,
      rowKeySelector,
      tabs,
      searchable,
      paginated,
      handlesPagination,
      handlesSorting,
      itemPathPrefix,
      pathname,
      actionItems,
      entity,
      searchPlaceholderMessage,
      searchQueryMaxLength,
      clickable,
    } = this.props
    const { page, query, tab, order, initialFetch, error } = this.state
    let orderDirection, orderBy

    // Parse order string.
    if (typeof order === 'string') {
      orderDirection = typeof order === 'string' && order[0] === '-' ? 'desc' : 'asc'
      orderBy = typeof order === 'string' && order[0] === '-' ? order.substr(1) : order
    }

    const filtersCls = classnames(style.filters, {
      [style.topRule]: tabs.length > 0,
    })

    // Disable sorting when incoming data was long enough to be paginated.
    const canHandleSorting = totalCount <= pageSize
    const disableSorting = handlesSorting && !canHandleSorting
    const handleSorting = handlesSorting && canHandleSorting && orderBy !== undefined

    const preparedItems = handleSorting
      ? lodashOrderBy(items, i => getByPath(i, orderBy), [orderDirection])
      : items

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
                maxLength={searchQueryMaxLength}
              />
            )}
            {(Boolean(actionItems) || mayAdd) && (
              <div className={style.actionItems}>
                {actionItems}
                {mayAdd && (
                  <Button.Link
                    primary
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
              details={error}
              noIngest
            />
          )}
          <Tabular
            paginated={paginated}
            page={page}
            totalCount={totalCount}
            pageSize={pageSize}
            onPageChange={this.onPageChange}
            loading={fetching}
            headers={headers}
            rowKeySelector={rowKeySelector}
            rowHrefSelector={mayLink ? this.rowHrefSelector : undefined}
            data={initialFetch ? [] : preparedItems}
            emptyMessage={sharedMessages.noMatch}
            handlesPagination={handlesPagination}
            onSortRequest={this.onOrderChange}
            order={orderDirection}
            orderBy={orderBy}
            clickable={clickable}
            disableSorting={disableSorting}
          />
        </Overlay>
      </div>
    )
  }
}

export default FetchTable
