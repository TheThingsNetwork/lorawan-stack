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
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import WithRootClass from '@ttn-lw/lib/components/with-root-class'

import ApplicationEvents from '@console/containers/application-events'

import Require from '@console/lib/components/require'

import style from '@console/views/app/app.styl'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayViewApplicationEvents } from '@console/lib/feature-checks'

const m = defineMessages({
  appData: 'Application data',
})

const ApplicationData = () => {
  const { appId } = useParams()

  useBreadcrumbs(
    'apps.single.data',
    <Breadcrumb path={`/applications/${appId}/data`} content={sharedMessages.liveData} />,
  )

  return (
    <Require
      featureCheck={mayViewApplicationEvents}
      otherwise={{ redirect: `/applications/${appId}` }}
    >
      <WithRootClass className={style.stageFlex} id="stage">
        <PageTitle hideHeading title={m.appData} />
        <ApplicationEvents appId={appId} />
      </WithRootClass>
    </Require>
  )
}

export default ApplicationData
