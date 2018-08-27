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
import bind from 'autobind-decorator'
import { storiesOf } from '@storybook/react'
import { withInfo } from '@storybook/addon-info'

import Button from '../../button'
import Tabs from '../../tabs'
import Input from '../../input'
import orders from '../orders'
import Table from '../'
import style from './story.styl'
import examples from './storyData'

const PAGE_SIZE = 5

const getNextOrder = function (order) {
  switch (order) {
  case orders.DEFAULT:
    return orders.ASCENDING
  case orders.ASCENDING:
    return orders.DESCENDING
  default:
    return orders.DEFAULT
  }
}

@bind
class Example extends Component {

  constructor (props) {
    super(props)

    this.state = {
      totalSize: props.data.length,
      dataShown: props.data.slice(0, PAGE_SIZE),
      page: 0,
      order: orders.DEFAULT,
      orderedBy: undefined,
    }
  }

  onSort (columnName) {
    const {
      order,
      orderedBy,
    } = this.state

    if (!!order && !!orderedBy && columnName !== orderedBy) {
      this.onDataRequest(0, getNextOrder(orders.DEFAULT), columnName)
      return
    }

    const newOrder = getNextOrder(order)
    const newOrderedBy = newOrder === orders.DEFAULT ? undefined : columnName
    this.onDataRequest(0, newOrder, newOrderedBy)
  }

  onPageChange (page) {
    const {
      order,
      orderedBy,
    } = this.state

    this.onDataRequest(page, order, orderedBy)
  }

  componentDidUpdate (prevProps) {
    const prevLength = prevProps.data.length
    const newLength = this.props.data.length

    // not robust check, but for this specific example is sufficient
    if (prevLength !== newLength) {
      this.setState({
        totalSize: newLength,
        dataShown: this.props.data.slice(0, PAGE_SIZE),
        page: this.props.forcePage,
      })
    }
  }

  onDataRequest (page, order = orders.DEFAULT, orderedBy) {
    let customSort = null
    if (order === orders.ASCENDING) {
      customSort = (a, b) => (
        a[orderedBy] > b[orderedBy]
          ? 1
          : a[orderedBy] < b[orderedBy]
            ? -1
            : 0
      )
    } else if (order === orders.DESCENDING) {
      customSort = (a, b) => (
        a[orderedBy] < b[orderedBy]
          ? 1
          : a[orderedBy] > b[orderedBy]
            ? -1
            : 0
      )
    } else {
      customSort = () => 0
    }

    const offset = page * PAGE_SIZE
    const { data } = this.props

    this.setState({
      dataShown: [].concat(data)
        .sort(customSort)
        .slice(offset, offset + PAGE_SIZE),
      order,
      orderedBy,
      page,
    })
  }

  render () {
    const {
      totalSize,
      dataShown,
      page,
      order,
      orderedBy,
    } = this.state

    const {
      headers,
      ...rest
    } = this.props

    return (
      <Table
        {...rest}
        pageCount={Math.ceil(totalSize / PAGE_SIZE)}
        onPageChange={this.onPageChange}
        onSortByColumn={this.onSort}
        headers={headers}
        rows={dataShown}
        emptyMessage="You have no applications"
        page={page}
        order={order}
        orderedBy={orderedBy}
      />
    )
  }
}

@bind
class TabbedExample extends Component {

  constructor (props) {
    super(props)

    const { tab, data } = props

    this.state = {
      tab,
      dataShown: data.filter(d => d.tabs.includes(tab)),
    }
  }

  onTabChange (tab) {
    const { data } = this.props
    this.setState({
      tab,
      dataShown: data.filter(d => d.tabs.includes(tab)),
    })
  }

  render () {
    const { tab, dataShown } = this.state
    const { onTabChange } = this
    const { headers, tabs } = this.props

    return (
      <div >
        <div className={style.wrapperHeader}>
          <Tabs
            className={style.tabs}
            tabs={tabs}
            onTabChange={onTabChange}
            active={tab}
          />
          <div className={style.search}>
            <Input className={style.searchInput} icon="search" placeholder="Applications" />
            <Button className={style.searchButton} message="Add Application" icon="add" />
          </div>
        </div>
        <Example
          forcePage={0}
          headers={headers}
          data={dataShown}
        />
      </div>
    )
  }
}

storiesOf('Table', module)
  .addDecorator((story, context) => withInfo({
    inline: true,
    header: false,
    source: false,
    propTables: [ Table ],
    propTablesExclude: [ Example ],
  })(story)(context))
  .add('Default', () => (
    <Example
      headers={examples.defaultExample.headers}
      data={examples.defaultExample.rows}
    />
  )).add('Loading', () => (
    <Example
      loading
      headers={examples.loadingExample.headers}
      data={examples.loadingExample.rows}
    />
  )).add('Custom cell', () => (
    <Example
      headers={examples.customCellExample.headers}
      data={examples.customCellExample.rows}
      small
    />
  )).add('Sortable', () => (
    <Example
      headers={examples.sortableExample.headers}
      data={examples.sortableExample.rows}
    />
  )).add('Empty', () => (
    <Example
      headers={examples.emptyExample.headers}
      data={examples.emptyExample.rows}
    />
  )).add('With custom wrapper', function () {
    return (
      <TabbedExample
        headers={examples.customWrapperExample.headers}
        data={examples.customWrapperExample.rows}
        tabs={[{ title: 'All', name: 'all' }, { title: 'Starred', name: 'starred' }]}
        tab={'all'}
      />
    )
  })
