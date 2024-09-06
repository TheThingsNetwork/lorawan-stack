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

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import OrganizationsTable from '@console/containers/organizations-table'

import sharedMessages from '@ttn-lw/lib/shared-messages'

const List = () => {
  useBreadcrumbs(
    'overview.orgs.list',
    <Breadcrumb path={`/organizations`} content={sharedMessages.list} />,
  )

  return (
    <div className="container container--xxl p-0">
      <IntlHelmet title={sharedMessages.organizations} />
      <OrganizationsTable />
    </div>
  )
}

export default List
