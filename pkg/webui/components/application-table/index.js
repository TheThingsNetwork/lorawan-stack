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

import React, { Component } from 'react'
import { connect } from 'react-redux'
import Query from 'query-string'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import { push } from 'connected-react-router'

import Tabs from '../tabs'
import Tabular from '../table'
import Button from '../button'
import Input from '../input'

import sharedMessages from '../../lib/shared-messages'
import debounce from '../../lib/debounce'

import {
  getApplicationsList,
  searchApplicationsList,
} from '../../actions/applications'

import style from './applications-table.styl'

const m = defineMessages({
  all: 'All',
  appId: 'Application ID',
  desc: 'Description',
  empty: 'No items matched your criteria',
  add: 'Add Application',
})

const tabs = [
  {
    title: m.all,
    name: 'all',
    disabled: true,
  },
]

const headers = [
  {
    name: 'application_id',
    displayName: m.appId,
  },
  {
    name: 'description',
    displayName: m.desc,
  },
]

const PAGE_SIZE = 3

@connect(({ applications }) => ({
  applications: applications.applications,
  totalCount: applications.totalCount,
  fetching: applications.fetching,
  fetchingSearch: applications.fetchingSearch,
  error: applications.error,
}))
@bind
export default class ApplicationTable extends Component {

  constructor (props) {
    super(props)

    this.state = {
      query: '',
    }

    this.requestSearch = debounce(this.requestSearch, 350)
  }

  getCurrentFilters () {
    const { search } = this.props.location
    const { page, tab, order, orderBy } = Query.parse(search)

    const result = {
      page: Number(page),
      tab,
      order,
      orderBy,
    }

    return result
  }

  requestSearch (query) {
    const { dispatch } = this.props
    const filters = this.getCurrentFilters()
    filters.query = query
    dispatch(searchApplicationsList(filters))
  }

  onQueryChange (query) {
    this.setState({ query })
    this.requestSearch(query)
  }

  onPageChange (page) {
    const { dispatch } = this.props
    const filters = this.getCurrentFilters()
    filters.page = page

    dispatch(push(`/console/applications?${Query.stringify(filters)}`))
  }

  onOrderChange (order, orderBy) {
    const { dispatch } = this.props
    const filters = this.getCurrentFilters()
    filters.order = order
    filters.orderBy = orderBy

    dispatch(push(`/console/applications?${Query.stringify(filters)}`))
  }

  onTabChange (tab) {
    const { dispatch } = this.props
    const filters = this.getCurrentFilters()
    filters.tab = tab

    dispatch(push(`/console/applications?${Query.stringify(filters)}`))
  }

  onApplicationClick (index) {
    const { applications, dispatch, match } = this.props
    const appId = applications[index].application_id

    dispatch(push(`${match.url}/${appId}`))
  }

  onApplicationAdd () {
    const { dispatch, match } = this.props

    dispatch(push(`${match.url}/add`))
  }

  componentDidMount () {
    this.fetchApplications(this.props.dispatch)
  }

  componentDidUpdate (newProps, newState) {
    if (this.props.location.search !== newProps.location.search) {
      this.fetchApplications(this.props.dispatch)
    }
  }

  fetchApplications (dispatch) {
    const filters = this.getCurrentFilters()
    if (filters.query) {
      dispatch(searchApplicationsList(filters))
    } else {
      dispatch(getApplicationsList(filters))
    }
  }

  render () {
    const {
      applications,
      totalCount,
      error,
      fetching,
      fetchingSearch,
    } = this.props

    const { query } = this.state
    const {
      tab = 'all',
      page = 1,
      ...rest
    } = this.getCurrentFilters()

    if (error) {
      return (
        <span>ERROR</span>
      )
    }

    const apps = applications.map(app => ({ ...app, clickable: true }))

    return (
      <React.Fragment>
        <div className={style.filters}>
          <div className={style.filterLeft}>
            <Tabs
              active={tab}
              className={style.tabs}
              tabs={tabs}
              onTabChange={this.onTabChange}
            />
          </div>
          <div className={style.filtersRight}>
            <Input
              value={query}
              icon="search"
              loading={fetchingSearch}
              onChange={this.onQueryChange}
            />
            <Button
              onClick={this.onApplicationAdd}
              className={style.addButton}
              message={m.add}
              icon="add"
            />
          </div>
        </div>
        <Tabular
          {...rest}
          paginated
          initialPage={page}
          page={page}
          totalCount={totalCount}
          pageSize={PAGE_SIZE}
          loading={fetching}
          onSortRequest={this.onOrderChange}
          onRowClick={this.onApplicationClick}
          onPageChange={this.onPageChange}
          headers={headers}
          data={apps}
          emptyMessage={sharedMessages.noMatch}
        />
      </React.Fragment>
    )
  }
}
