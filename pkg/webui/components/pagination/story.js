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

import React, { useState } from 'react'

import Pagination from '.'

export default {
  title: 'Pagination',
  component: Pagination,
}

export const Default = () => {
  const [pageSize, setPageSize] = useState(10)

  return (
    <div>
      <Pagination
        pageCount={3}
        marginPagesDisplayed={2}
        pageSize={pageSize}
        setPageSize={setPageSize}
      />
    </div>
  )
}

export const AllPagesWithoutGaps = () => {
  const [pageSize, setPageSize] = useState(10)

  return (
    <Pagination
      pageCount={10}
      pageRangeDisplayed={10}
      pageSize={pageSize}
      setPageSize={setPageSize}
    />
  )
}

AllPagesWithoutGaps.story = {
  name: 'All pages (without gaps)',
}

export const WithGaps = () => {
  const [pageSize, setPageSize] = useState(10)

  return (
    <div>
      <Pagination
        pageCount={20}
        marginPagesDisplayed={2}
        pageSize={pageSize}
        setPageSize={setPageSize}
      />
      <Pagination
        pageCount={9}
        pageRangeDisplayed={1}
        marginPagesDisplayed={2}
        pageSize={pageSize}
        setPageSize={setPageSize}
      />
      <Pagination
        pageCount={9}
        pageRangeDisplayed={3}
        marginPagesDisplayed={2}
        pageSize={pageSize}
        setPageSize={setPageSize}
      />
    </div>
  )
}

WithGaps.story = {
  name: 'With gaps',
}
