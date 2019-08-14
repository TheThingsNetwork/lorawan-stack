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
import { storiesOf } from '@storybook/react'
import { withInfo } from '@storybook/addon-info'
import { action } from '@storybook/addon-actions'

import doc from '../table.md'

import Tabular from '../'
import examples from './storyData'

@bind
class LoadingExample extends React.Component {
  state = {
    loading: true,
  }

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

@bind
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

  getDelay(slow) {
    if (!slow) {
      return Math.floor(Math.random() * (450 - 100)) + 100
    }

    return Math.floor(Math.random() * (3000 - 1000)) + 1000
  }

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

@bind
class ClickableExample extends React.Component {
  state = {
    clicked: null,
  }

  onRowClick(rowIndex) {
    action('onRowClick')({ index: rowIndex, id: this.props.data[rowIndex].appId })
    // push to history if link functionality required
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

@bind
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

  onSort(order, orderBy) {
    // this.setState({ order, orderBy })
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

storiesOf('Table/Tabular', module)
  .addDecorator((story, context) =>
    withInfo({
      inline: true,
      header: false,
      source: false,
      propTables: [Tabular],
      propTablesExclude: [LoadingExample, PaginatedExample, ClickableExample],
      text: doc,
    })(story)(context),
  )
  .add('Default', () => (
    <Tabular
      data={examples.defaultExample.rows}
      headers={examples.defaultExample.headers}
      emptyMessage="No entries to display"
    />
  ))
  .add('Loading (slow)', () => (
    <LoadingExample
      slow
      data={examples.defaultExample.rows}
      headers={examples.defaultExample.headers}
      emptyMessage="No entries to display"
    />
  ))
  .add('Empty', () => (
    <Tabular
      data={examples.emptyExample.rows}
      headers={examples.emptyExample.headers}
      emptyMessage="No entries to display"
    />
  ))
  .add('Paginated (slow loading)', () => (
    <PaginatedExample
      slow
      data={examples.paginatedExample.rows}
      headers={examples.paginatedExample.headers}
      emptyMessage="No entries to display"
    />
  ))
  .add('Paginated (fast loading)', () => (
    <PaginatedExample
      fast
      data={examples.paginatedExample.rows}
      headers={examples.paginatedExample.headers}
      emptyMessage="No entries to display"
    />
  ))
  .add('Small', () => (
    <PaginatedExample
      small
      fast
      data={examples.defaultExample.rows}
      headers={examples.defaultExample.headers}
      emptyMessage="No entries to display"
    />
  ))
  .add('Clickable rows', () => (
    <ClickableExample
      data={examples.clickableRowsExample.rows}
      headers={examples.clickableRowsExample.headers}
      emptyMessage="No entries to display"
    />
  ))
  .add('Sortable', () => (
    <SortableExample
      data={examples.sortableExample.rows}
      headers={examples.sortableExample.headers}
      emptyMessage="No entries to display"
    />
  ))
