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
import { useParams } from 'react-router-dom'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import GatewayLocationForm from '@console/containers/gateway-location-form'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayViewOrEditGatewayLocation } from '@console/lib/feature-checks'

const GatewayLocation = () => {
  const { gtwId } = useParams()

  useBreadcrumbs(
    'gtws.single.data',
    <Breadcrumb path={`/gateways/${gtwId}/location`} content={sharedMessages.location} />,
  )

  return (
    <Require
      featureCheck={mayViewOrEditGatewayLocation}
      otherwise={{ redirect: `/gateways/${gtwId}` }}
    >
      <div className="container container--xl grid">
        <div className="item-12">
          <PageTitle title={sharedMessages.location} />
          <GatewayLocationForm />
        </div>
      </div>
    </Require>
  )
}

export default GatewayLocation
