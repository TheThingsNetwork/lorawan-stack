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

import Pagination from '.'

export default {
  title: 'Pagination',
}

export const Default = () => (
  <div>
    <Pagination pageCount={1} />
    <Pagination pageCount={3} initialPage={2} marginPagesDisplayed={2} />
    <Pagination pageCount={3} initialPage={3} marginPagesDisplayed={2} />
  </div>
)

export const AllPagesWithoutGaps = () => <Pagination pageCount={10} pageRangeDisplayed={10} />

AllPagesWithoutGaps.story = {
  name: 'All pages (without gaps)',
}

export const WithGaps = () => (
  <div>
    <Pagination pageCount={20} marginPagesDisplayed={2} />
    <Pagination pageCount={9} initialPage={4} pageRangeDisplayed={1} marginPagesDisplayed={2} />
    <Pagination pageCount={9} initialPage={4} pageRangeDisplayed={3} marginPagesDisplayed={2} />
  </div>
)

WithGaps.story = {
  name: 'With gaps',
}
