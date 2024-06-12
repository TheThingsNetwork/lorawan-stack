// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import sharedMessages from '@ttn-lw/lib/shared-messages'

const m = defineMessages({
  theThingsStationConnectionProfiles: 'The Things Station connection profiles',
})

const GatewayConnectionProfiles = () => {
  const { gtwId } = useParams()

  useBreadcrumbs(
    'gtws.single.the-things-station.connection-profiles',
    <Breadcrumb
      path={`/gateways/${gtwId}/the-things-station/connection-profiles`}
      content={sharedMessages.connectionProfiles}
    />,
  )

  return (
    <>
      <PageTitle title={m.theThingsStationConnectionProfiles} />
    </>
  )
}

export default GatewayConnectionProfiles
