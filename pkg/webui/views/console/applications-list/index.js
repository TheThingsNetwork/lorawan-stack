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
import bind from 'autobind-decorator'
import { connect } from 'react-redux'
import { push } from 'connected-react-router'
import Query from 'query-string'
import { Row, Col } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import Tabs from '../../../components/tabs'
import Tabular from '../../../components/table'
import Button from '../../../components/button'
import Input from '../../../components/input'
import Message from '../../../lib/components/message'
import sharedMessages from '../../../lib/shared-messages'
import IntlHelmet from '../../../lib/components/intl-helmet'

import {
  getApplicationsList,
  searchApplicationsList,
  changeApplicationsPage,
  changeApplicationsOrder,
  changeApplicationsTab,
  changeApplicationsSearch,
} from '../../../actions/applications'

import style from './applications-list.styl'

const m = defineMessages({
  addApp: 'Add Application',
})

const tabs = [
  {
    title: sharedMessages.all,
    name: 'all',
    disabled: true,
  },
]

const headers = [
  {
    name: 'application_id',
    displayName: sharedMessages.appId,
  },
  {
    name: 'description',
    displayName: sharedMessages.description,
  },
]

const PAGE_SIZE = 3

@connect(({ applications }) => ({
  applications: applications.applications,
  totalCount: applications.totalCount,
  fetching: applications.fetching,
  fetchingSearch: applications.fetchingSearch,
  error: applications.error,
  filters: applications.filters,
}))
@bind
export default class List extends React.Component {

  onQueryChange (query) {
    const { dispatch } = this.props

    dispatch(changeApplicationsSearch(query))
  }

  onPageChange (page) {
    const { dispatch } = this.props

    dispatch(changeApplicationsPage(page))
  }

  onOrderChange (order, orderBy) {
    const { dispatch } = this.props

    dispatch(changeApplicationsOrder(order, orderBy))
  }

  onTabChange (tab) {
    const { dispatch } = this.props

    dispatch(changeApplicationsTab(tab))
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
    const { filters, location, dispatch } = this.props

    // process query params first and only after consider using default props
    const queryParams = Query.parse(location.search)

    const page = Number(queryParams.page) || filters.page
    const order = queryParams.order || filters.order
    const orderBy = queryParams.orderBy || filters.orderBy
    const tab = queryParams.tab || filters.tab
    const query = queryParams.query || filters.query

    if (query) {
      dispatch(searchApplicationsList(page, order, orderBy, tab, query))
    } else {
      dispatch(getApplicationsList(page, order, orderBy, tab))
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

    const { query, page, tab, ...filters } = this.props.filters

    if (error) {
      return (
        <span>ERROR</span>
      )
    }

    const apps = applications.map(app => ({ ...app, clickable: true }))

    return (
      <Row>
        <IntlHelmet>
          <title><Message content={sharedMessages.applications} /></title>
        </IntlHelmet>
        <Col sm={12}>
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
                value={query || ''}
                icon="search"
                loading={fetchingSearch}
                onChange={this.onQueryChange}
              />
              <Button
                onClick={this.onApplicationAdd}
                className={style.addButton}
                message={m.addApp}
                icon="add"
              />
            </div>
          </div>
          <Tabular
            {...filters}
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
        </Col>
      </Row>
    )
  }
}
