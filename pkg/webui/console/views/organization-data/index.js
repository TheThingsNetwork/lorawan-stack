// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
import { defineMessages } from 'react-intl'
import { useParams } from 'react-router-dom'

import PageTitle from '@ttn-lw/components/page-title'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import WithRootClass from '@ttn-lw/lib/components/with-root-class'

import OrganizationEvents from '@console/containers/organization-events'

import style from '@console/views/app/app.styl'

import sharedMessages from '@ttn-lw/lib/shared-messages'

const m = defineMessages({
  orgData: 'Organization data',
})

const Data = () => {
  const { orgId } = useParams()

  useBreadcrumbs('orgs.single.data', [
    {
      path: `/organizations/${orgId}/data`,
      content: sharedMessages.liveData,
    },
  ])

  return (
    <WithRootClass className={style.stageFlex} id="stage">
      <PageTitle hideHeading title={m.orgData} />
      <OrganizationEvents orgId={orgId} />
    </WithRootClass>
  )
}

export default Data
