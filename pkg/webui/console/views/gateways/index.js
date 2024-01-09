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
import { Routes, Route } from 'react-router-dom'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumbs from '@ttn-lw/components/breadcrumbs'

import ValidateRouteParam from '@ttn-lw/lib/components/validate-route-param'

import Require from '@console/lib/components/require'

import Gateway from '@console/views/gateway'
import GatewayAdd from '@console/views/gateway-add'
import GatewaysList from '@console/views/gateways-list'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { pathId as pathIdRegexp } from '@ttn-lw/lib/regexp'

import { mayViewGateways } from '@console/lib/feature-checks'

const Gateways = () => {
  useBreadcrumbs('gtws', <Breadcrumb path="/gateways" content={sharedMessages.gateways} />)
  return (
    <Require featureCheck={mayViewGateways} otherwise={{ redirect: '/' }}>
      <Breadcrumbs />
      <Routes>
        <Route index Component={GatewaysList} />
        <Route path="add" Component={GatewayAdd} />
        <Route
          path={`:gtwId/*`}
          element={<ValidateRouteParam check={{ gtwId: pathIdRegexp }} Component={Gateway} />}
        />
      </Routes>
    </Require>
  )
}

export default Gateways
