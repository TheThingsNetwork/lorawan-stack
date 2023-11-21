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

/* eslint-disable react/prop-types */

import React from 'react'
import bind from 'autobind-decorator'
import { action } from '@storybook/addon-actions'

import doc from '../table.md'
import Tabular from '..'

import examples from './storyData'

class LoadingExample extends React.Component {
  state = {
    loading: true,
  }

  @bind
  toggleLoading() {
    this.setState(prev => ({
      loading: !prev.loading,
    }))
  }

  render() {
    const { loading } = this.state
    return (
      <div>
        <Tabular {...this.props} loading={loading} />
        <div style={{ textAlign: 'center' }}>
          <button style={{ marginTop: '20px' }} onClick={this.toggleLoading}>
            {loading ? 'Stop Loading' : 'Start Loading'}
          </button>
        </div>
      </div>
    )
  }
}

const PAGE_SIZE = 3

class PaginatedExample extends React.Component {
  constructor(props) {
    super(props)

    this.state = {
      page: 1,
      loading: true,
      data: [],
    }
  }

  componentDidMount() {
    this.requestNextPage(1)
  }

  componentWillUnmount() {
    window.clearTimeout(this.timeout)
  }

  @bind
  getDelay(slow) {
    if (!slow) {
      return Math.floor(Math.random() * (450 - 100)) + 100
    }

    return Math.floor(Math.random() * (3000 - 1000)) + 1000
  }

  @bind
  requestNextPage(page) {
    action('requestNextPage')(page)
    const offset = (page - 1) * PAGE_SIZE
    const delay = this.getDelay(this.props.slow)
    this.timeout = setTimeout(
      () =>
        this.setState({
          page,
          loading: false,
          data: this.props.data.slice(offset, offset + PAGE_SIZE),
        }),
      delay,
    )
  }

  @bind
  onPageChange(page) {
    this.setState({ loading: true }, () => this.requestNextPage(page))
  }

  render() {
    const { page, data, loading } = this.state
    const { small, headers, ...rest } = this.props

    return (
      <Tabular
        {...rest}
        paginated
        small={small}
        headers={headers}
        data={data}
        page={page}
        totalCount={this.props.data.length}
        pageSize={PAGE_SIZE}
        loading={loading}
        onPageChange={this.onPageChange}
      />
    )
  }
}

class ClickableExample extends React.Component {
  state = {
    clicked: null,
  }

  @bind
  onRowClick(rowIndex) {
    action('onRowClick')({ index: rowIndex, id: this.props.data[rowIndex].appId })
    // Push to history if link functionality required
    this.setState({ clicked: this.props.data[rowIndex].appId })
  }

  render() {
    const { clicked } = this.state

    return (
      <div>
        <Tabular {...this.props} onRowClick={this.onRowClick} />
        <span>
          You clicked on: <strong>{clicked ? clicked : 'nothing'}</strong>
        </span>
      </div>
    )
  }
}

class SortableExample extends React.Component {
  state = {
    order: undefined,
    orderBy: undefined,
    loading: false,
    data: [],
  }

  componentDidMount() {
    this.onSortRequest(undefined, undefined)
  }

  componentWillUnmount() {
    window.clearTimeout(this.timeout)
  }

  asc(a, b) {
    return a > b ? 1 : a < b ? -1 : 0
  }

  @bind
  onSort(order, orderBy) {
    // This.setState({ order, orderBy })
    const data = this.props.data
    const asc = this.asc
    action('onSort')({ order, orderBy })

    this.timeout = setTimeout(
      () =>
        this.setState({
          loading: false,
          data: []
            .concat(data)
            .sort((a, b) =>
              order === 'asc'
                ? asc(a[orderBy], b[orderBy])
                : order === 'desc'
                ? -asc(a[orderBy], b[orderBy])
                : 0,
            ),
        }),
      800,
    )
  }

  @bind
  onSortRequest(order, orderBy) {
    this.setState({ loading: true, order, orderBy }, () => this.onSort(order, orderBy))
  }

  render() {
    const { order, orderBy, loading } = this.state
    const { data, ...rest } = this.props

    return (
      <Tabular
        {...rest}
        order={order}
        orderBy={orderBy}
        data={this.state.data}
        loading={loading}
        onSortRequest={this.onSortRequest}
      />
    )
  }
}

export default {
  title: 'Table/Tabular',
  component: Tabular,
  parameters: {
    docs: {
      description: {
        component: doc,
      },
    },
  },
}

export const Default = () => (
  <Tabular
    data={examples.defaultExample.rows}
    headers={examples.defaultExample.headers}
    emptyMessage="No entries to display"
  />
)

export const LoadingSlow = () => (
  <LoadingExample
    slow
    data={examples.defaultExample.rows}
    headers={examples.defaultExample.headers}
    emptyMessage="No entries to display"
  />
)

LoadingSlow.story = {
  name: 'Loading (slow)',
}

export const Empty = () => (
  <Tabular
    data={examples.emptyExample.rows}
    headers={examples.emptyExample.headers}
    emptyMessage="No entries to display"
  />
)

export const PaginatedSlowLoading = () => (
  <PaginatedExample
    slow
    data={examples.paginatedExample.rows}
    headers={examples.paginatedExample.headers}
    emptyMessage="No entries to display"
  />
)

PaginatedSlowLoading.story = {
  name: 'Paginated (slow loading)',
}

export const PaginatedFastLoading = () => (
  <PaginatedExample
    fast
    data={examples.paginatedExample.rows}
    headers={examples.paginatedExample.headers}
    emptyMessage="No entries to display"
  />
)

PaginatedFastLoading.story = {
  name: 'Paginated (fast loading)',
}

export const Small = () => (
  <PaginatedExample
    small
    fast
    data={examples.defaultExample.rows}
    headers={examples.defaultExample.headers}
    emptyMessage="No entries to display"
  />
)

export const ClickableRows = () => (
  <ClickableExample
    data={examples.clickableRowsExample.rows}
    headers={examples.clickableRowsExample.headers}
    emptyMessage="No entries to display"
  />
)

ClickableRows.story = {
  name: 'Clickable rows',
}

export const Sortable = () => (
  <SortableExample
    data={examples.sortableExample.rows}
    headers={examples.sortableExample.headers}
    emptyMessage="No entries to display"
  />
)
