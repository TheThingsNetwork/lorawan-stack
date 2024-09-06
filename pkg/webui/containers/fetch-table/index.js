// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useEffect, useState, useRef } from 'react'
import { defineMessages } from 'react-intl'
import { useDispatch, useSelector } from 'react-redux'
import { orderBy as lodashOrderBy } from 'lodash'
import classNames from 'classnames'

import { PAGE_SIZES } from '@ttn-lw/constants/page-sizes'

import { IconPlus } from '@ttn-lw/components/icon'
import Tabular from '@ttn-lw/components/table'
import Input from '@ttn-lw/components/input'
import Button from '@ttn-lw/components/button'
import Tabs from '@ttn-lw/components/tabs'
import Overlay from '@ttn-lw/components/overlay'
import ErrorNotification from '@ttn-lw/components/error-notification'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import getByPath from '@ttn-lw/lib/get-by-path'
import useDebounce from '@ttn-lw/lib/hooks/use-debounce'
import useQueryState from '@ttn-lw/lib/hooks/use-query-state'

import { selectPageSize } from '@console/store/selectors/user-preferences'

import style from './fetch-table.styl'

const DEFAULT_PAGE = 1

const pageValidator = page => (!Boolean(page) || page < 0 ? DEFAULT_PAGE : page)
const orderValidator = order =>
  typeof order === 'string' && order.match(/-?[a-z0-9]/) === null ? undefined : order

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

const FetchTable = props => {
  const {
    pageSize: initialPageSize,
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
    actionItems,
    entity,
    searchPlaceholderMessage,
    searchQueryMaxLength,
    clickable,
    defaultOrder,
    getItemPathPrefix,
    searchItemsAction,
    getItemsAction,
    baseDataSelector,
    className,
    headerClassName,
    panelStyle,
    filtersClassName,
  } = props

  const globalPageSize = useSelector(selectPageSize)
  const [pageSize, setPageSize] = useState(initialPageSize)
  const isMounted = useRef(true)
  const dispatch = useDispatch()
  const defaultTab = tabs.length > 0 ? tabs[0].name : undefined
  const [page, setPage] = useQueryState('page', 1, parseInt)
  const [tab, setTab] = useQueryState('tab', defaultTab)
  const [order, setOrder] = useQueryState('order', defaultOrder)
  const [query, setQuery] = useQueryState('query', '')
  const debouncedQuery = useDebounce(
    query,
    350,
    useCallback(() => {
      setPage(1)
    }, [setPage]),
  )
  const [initialFetch, setInitialFetch] = useState(true)
  const base = useSelector(state => baseDataSelector(state, props))
  const items = base[props.entity] || []
  const totalCount = base.totalCount || 0
  const mayAdd = 'mayAdd' in base ? base.mayAdd : true
  const mayLink = 'mayLink' in base ? base.mayLink : true

  const filters = { query: debouncedQuery, tab, order, page, limit: globalPageSize ?? pageSize }
  const [fetching, setFetching] = useState(true)
  const [error, setError] = useState(undefined)
  let orderDirection, orderBy
  // Parse order string.
  if (typeof order === 'string') {
    orderDirection = typeof order === 'string' && order[0] === '-' ? 'desc' : 'asc'
    orderBy = typeof order === 'string' && order[0] === '-' ? order.substr(1) : order
  }
  // Disable sorting when incoming data was long enough to be paginated.
  const canHandleSorting = totalCount <= globalPageSize
  const disableSorting = handlesSorting && !canHandleSorting
  const handleSorting = handlesSorting && canHandleSorting && orderBy !== undefined
  if (!handleSorting) {
    filters.order = order
  }

  useEffect(
    () => () => {
      isMounted.current = false
    },
    [],
  )

  useEffect(() => {
    const fetchItems = async () => {
      setFetching(true)
      const f = { query: debouncedQuery || '', page, limit: globalPageSize ?? pageSize }
      if (tabs.find(t => t.name === tab)) {
        f.tab = tab
      } else {
        setTab(defaultTab)
        f.tab = undefined
      }

      if (orderValidator(order)) {
        f.order = order
      } else {
        if (isMounted.current) {
          setOrder(defaultOrder)
        }
        f.order = defaultOrder
      }

      try {
        if (f.query && searchItemsAction) {
          await dispatch(attachPromise(searchItemsAction(f)))
        } else {
          await dispatch(attachPromise(getItemsAction(f)))
        }
        if (isMounted.current) {
          setFetching(false)
          setInitialFetch(false)
        }
      } catch (error) {
        if (isMounted.current) {
          setError(error)
          setFetching(false)
        }
      }
    }
    fetchItems()
  }, [
    debouncedQuery,
    defaultOrder,
    defaultTab,
    dispatch,
    getItemsAction,
    order,
    page,
    globalPageSize,
    pageSize,
    searchItemsAction,
    setOrder,
    setTab,
    tab,
    tabs,
  ])

  const onPageChange = useCallback(
    page => {
      setPage(pageValidator(page))
    },
    [setPage],
  )

  const onQueryChange = useCallback(
    query => {
      setQuery(query)
    },
    [setQuery],
  )

  const onOrderChange = useCallback(
    (order, orderBy) => {
      const filterOrder = `${order === 'desc' ? '-' : ''}${orderBy}`

      setOrder(filterOrder)
    },
    [setOrder],
  )

  const onTabChange = useCallback(
    tab => {
      setTab(tab)
      setPage(1)
      setQuery('')
    },
    [setPage, setQuery, setTab],
  )

  const rowHrefSelector = useCallback(
    item => {
      const entitySingle = entity.substr(0, entity.length - 1)
      let entityPath
      if (Boolean(getItemPathPrefix)) {
        entityPath = getItemPathPrefix(item)
      } else {
        const item_id = item.id || item.ids[`${entitySingle}_id`]
        entityPath = `${itemPathPrefix}${item_id}`
      }

      return entityPath
    },
    [entity, getItemPathPrefix, itemPathPrefix],
  )

  const preparedItems = handleSorting
    ? lodashOrderBy(items, i => getByPath(i, orderBy), [orderDirection])
    : items

  const filtersCls = classNames(filtersClassName, style.filters, {
    [style.topRule]: tabs.length > 0,
  })

  // Go back to page 1 when no items are left on the current page.
  useEffect(() => {
    if (preparedItems.length === 0 && page > 1 && !initialFetch) {
      setPage(1)
    }
  }, [initialFetch, page, preparedItems.length, setPage])

  return (
    <div data-test-id={`${entity}-table`}>
      {(tabs.length > 0 || tableTitle || actionItems || mayAdd || searchable) && (
        <div className={filtersCls}>
          <div className={style.filtersLeft}>
            {tabs.length > 0 ? (
              <Tabs
                active={tab}
                className={style.tabs}
                tabs={tabs}
                onTabChange={onTabChange}
                toggleStyle
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
                onChange={onQueryChange}
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
                    icon={IconPlus}
                    to={`${itemPathPrefix}add`}
                  />
                )}
              </div>
            )}
          </div>
        </div>
      )}
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
          pageSize={parseInt(globalPageSize ?? pageSize)}
          setPageSize={setPageSize}
          onPageChange={onPageChange}
          loading={fetching}
          headers={headers}
          rowKeySelector={rowKeySelector}
          rowHrefSelector={mayLink ? rowHrefSelector : undefined}
          data={initialFetch ? [] : preparedItems}
          emptyMessage={sharedMessages.noMatch}
          handlesPagination={handlesPagination}
          onSortRequest={onOrderChange}
          order={orderDirection}
          orderBy={orderBy}
          clickable={clickable}
          disableSorting={disableSorting}
          className={className}
          headerClassName={headerClassName}
          panelStyle={panelStyle}
        />
      </Overlay>
    </div>
  )
}

FetchTable.propTypes = {
  actionItems: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]),
  addMessage: PropTypes.message,
  baseDataSelector: PropTypes.func.isRequired,
  className: PropTypes.string,
  clickable: PropTypes.bool,
  defaultOrder: PropTypes.string,
  entity: PropTypes.string.isRequired,
  filtersClassName: PropTypes.string,
  getItemPathPrefix: PropTypes.func,
  getItemsAction: PropTypes.func.isRequired,
  handlesPagination: PropTypes.bool,
  handlesSorting: PropTypes.bool,
  headerClassName: PropTypes.string,
  headers: PropTypes.arrayOf(
    PropTypes.shape({
      displayName: PropTypes.message.isRequired,
      getValue: PropTypes.func,
      name: PropTypes.string,
      render: PropTypes.func,
      align: PropTypes.oneOf(['left', 'right', 'center']),
      sortable: PropTypes.bool,
      width: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    }),
  ),
  itemPathPrefix: PropTypes.string,
  pageSize: PropTypes.number,
  paginated: PropTypes.bool,
  panelStyle: PropTypes.bool,
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
      icon: PropTypes.icon,
      disabled: PropTypes.bool,
    }),
  ),
}

FetchTable.defaultProps = {
  className: undefined,
  getItemPathPrefix: undefined,
  searchItemsAction: undefined,
  pageSize: PAGE_SIZES.REGULAR,
  itemPathPrefix: '',
  searchable: false,
  searchPlaceholderMessage: sharedMessages.search,
  searchQueryMaxLength: 50,
  paginated: true,
  handlesPagination: false,
  handlesSorting: false,
  headers: [],
  rowKeySelector: undefined,
  addMessage: undefined,
  tableTitle: undefined,
  tabs: [],
  actionItems: null,
  clickable: true,
  defaultOrder: undefined,
  headerClassName: undefined,
  panelStyle: false,
  filtersClassName: undefined,
}

export default FetchTable
