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

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getApplicationsList } from '@console/store/actions/applications'
import { getGatewaysList } from '@console/store/actions/gateways'

const Overview = () => {
  useBreadcrumbs('overview', <Breadcrumb path="/" content={sharedMessages.overview} />)

  return (
    <RequireRequest requestAction={[getApplicationsList(), getGatewaysList()]}>
      <div className="grid p-ls-xs3 box-border">
        <div className="item-6 m:item-6 l:item-12 s:item-12">
          <div style={{ height: '30rem', backgroundColor: 'lightgray' }} />
        </div>
        <div className="item-6 m:item-6 l:item-12 s:item-12">
          <div style={{ height: '30rem', backgroundColor: 'lightgray' }} />
        </div>
        <div className="item-6 m:item-6 l:item-12 s:item-12">
          <div style={{ height: '30rem', backgroundColor: 'lightgray' }} />
        </div>
        <div className="item-6 m:item-6 l:item-12 s:item-12">
          <div style={{ height: '30rem', backgroundColor: 'lightgray' }} />
        </div>
      </div>
    </RequireRequest>
  )
}

export default Overview
