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
import { Route, Routes } from 'react-router-dom'
import { useSelector } from 'react-redux'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import ValidateRouteParam from '@ttn-lw/lib/components/validate-route-param'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'

import Require from '@console/lib/components/require'

import Device from '@console/views/device'
import DeviceImport from '@console/views/device-import'
import DeviceAdd from '@console/views/device-add'
import DeviceList from '@console/views/device-list'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { pathId as pathIdRegexp } from '@ttn-lw/lib/regexp'

import { mayViewApplicationDevices } from '@console/lib/feature-checks'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'

const Devices = () => {
  const appId = useSelector(selectSelectedApplicationId)

  useBreadcrumbs(
    'apps.single.devices',
    <Breadcrumb path={`/applications/${appId}/devices`} content={sharedMessages.devices} />,
  )

  return (
    <Require
      featureCheck={mayViewApplicationDevices}
      otherwise={{ redirect: `/applications/${appId}` }}
    >
      <Routes>
        <Route index Component={DeviceList} />
        <Route path="add" Component={DeviceAdd} />
        <Route path="import" Component={DeviceImport} />
        <Route
          path=":devId/*"
          element={<ValidateRouteParam check={{ devId: pathIdRegexp }} Component={Device} />}
        />
        <Route path="*" element={<GenericNotFound />} />
      </Routes>
    </Require>
  )
}

export default Devices
