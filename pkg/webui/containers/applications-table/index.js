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
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import { push } from 'connected-react-router'

import Tabs from '../../components/tabs'
import Tabular from '../../components/table'
import Button from '../../components/button'
import Input from '../../components/input'

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
  location,
}))
@bind
export default class ApplicationsTable extends Component {

  constructor (props) {
    super(props)

    this.state = {
      query: '',
      tab: 'all',
      page: 1,
      order: undefined,
      orderBy: undefined,
    }

    this.requestSearch = debounce(this.requestSearch, 350)
  }

  requestSearch (query) {
    const { dispatch } = this.props
    const filters = this.state
    filters.query = query
    filters.page = 1
    this.setState({ page: 1 })
    dispatch(searchApplicationsList(filters))
  }

  onQueryChange (query) {
    this.setState({ query })
    this.requestSearch(query)
  }

  async onPageChange (page) {
    const { dispatch } = this.props
    await this.setState({ page })
    this.fetchApplications(dispatch)
  }

  onOrderChange (order, orderBy) {
    const { dispatch } = this.props
    const filters = this.state
    filters.order = order
    filters.orderBy = orderBy

    this.setState(filters)
    this.fetchApplications(dispatch)
  }

  onTabChange (tab) {
    const { dispatch } = this.props
    const filters = this.state
    filters.tab = tab

    this.setState(filters)
    this.fetchApplications(dispatch)
  }

  onApplicationClick (index) {
    const { applications, dispatch, location } = this.props
    const appId = applications[index].application_id

    dispatch(push(`${location.pathname}/${appId}`))
  }

  onApplicationAdd () {
    const { dispatch, location } = this.props

    dispatch(push(`${location.pathname}/add`))
  }

  componentDidMount () {
    this.fetchApplications(this.props.dispatch)
  }

  fetchApplications (dispatch) {
    const filters = this.state
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

    const {
      page,
      tab,
      query,
    } = this.state

    if (error) {
      return <span>ERROR</span>
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
