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
import { connect } from 'react-redux'
import Query from 'query-string'

import { getApplicationsList } from '../../../actions/applications'

@connect(({ applications }) => ({
  applications: applications.applications,
  totalCount: applications.totalCount,
  fetching: applications.fetching,
  fetchingSearch: applications.fetchingSearch,
  error: applications.error,
  filters: applications.filters,
}))
export default class List extends React.Component {

  componentDidMount () {
    const {
      dispatch,
      location,
      page,
      order,
      orderBy,
    } = this.props

    const queryParams = Query.parse(location.search)

    const urlPage = queryParams.page || page
    const urlOrder = queryParams.order || order
    const urlOrderBy = queryParams.orderBy || orderBy

    dispatch(getApplicationsList(urlPage, urlOrder, urlOrderBy))
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
                message={m.add}
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
