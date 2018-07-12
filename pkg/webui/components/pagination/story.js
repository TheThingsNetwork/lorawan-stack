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

import Pagination from '.'

@bind
class Example extends Component {

  state = {
    currentPage: 0,
  }

  onPageChange (page) {
    this.setState({ currentPage: page.selected })
  }

  render () {
    return (
      <Pagination
        onPageChange={this.onPageChange}
        {...this.props}
      />
    )
  }
}

const Separator = () => (<div style={{ height: '20px' }} />)

storiesOf('Pagination', module)
  .add('Default', () =>
    (
      <div>
        <Example
          pageCount={1}
        />
        <Separator />
        <Example
          pageCount={3}
          initialPage={0}
          marginPagesDisplayed={2}
        />
        <Separator />
        <Example
          pageCount={3}
          initialPage={2}
          marginPagesDisplayed={2}
        />
      </div>
    )
  ).add('All pages (without gaps)', () =>
    (
      <Example
        pageCount={10}
        pageRangeDisplayed={10}
      />
    )
  ).add('With gaps', () =>
    (
      <div>
        <Separator />
        <Example
          pageCount={20}
          marginPagesDisplayed={2}
        />
        <Separator />
        <Example
          pageCount={9}
          initialPage={4}
          pageRangeDisplayed={1}
          marginPagesDisplayed={2}
        />
        <Separator />
        <Example
          pageCount={9}
          initialPage={4}
          pageRangeDisplayed={3}
          marginPagesDisplayed={2}
        />
      </div>
    ))
