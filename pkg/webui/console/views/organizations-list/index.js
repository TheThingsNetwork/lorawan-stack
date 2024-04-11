// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import { PAGE_SIZES } from '@ttn-lw/constants/page-sizes'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import OrganizationsTable from '@console/containers/organizations-table'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import getCookie from '@console/lib/table-utils'

const List = () => {
  const orgPageSize = getCookie('organizations-list-page-size')
  const orgParam = `?page-size=${orgPageSize ? orgPageSize : PAGE_SIZES.REGULAR}`
  useBreadcrumbs(
    'orgs.list',
    <Breadcrumb path={`/organizations${orgParam}`} content={sharedMessages.list} />,
  )

  return (
    <div className="container container--lg ">
      <IntlHelmet title={sharedMessages.organizations} />
      <OrganizationsTable />
    </div>
  )
}

export default List
