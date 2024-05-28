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

import React, { useCallback } from 'react'
import { defineMessages } from 'react-intl'
import { useParams } from 'react-router-dom'
import { useSelector } from 'react-redux'

import tts from '@console/api/tts'

import { IconFileDownload } from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import DataSheet from '@ttn-lw/components/data-sheet'
import toast from '@ttn-lw/components/toast'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import GatewayMap from '@console/components/gateway-map'
import BlurryNetworkActivityPanel from '@console/components/blurry-network-activity-panel'

import GatewayOverviewHeader from '@console/containers/gateway-overview-header'
import GatewayEvents from '@console/containers/gateway-events'
import GatewayTitleSection from '@console/containers/gateway-title-section'
import TopEntitiesDashboardPanel from '@console/containers/top-entities-dashboard-panel'
import NotificationsDashboardPanel from '@console/containers/notifications-dashboard-panel'
import DocumentationDashboardPanel from '@console/containers/documentation-dashboard-panel'
import ShortcutPanel from '@console/containers/shortcut-panel'

import Require from '@console/lib/components/require'

import { composeDataUri, downloadDataUriAsFile } from '@ttn-lw/lib/data-uri'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  mayViewGatewayInfo,
  mayViewGatewayConfJson,
  checkFromState,
} from '@console/lib/feature-checks'

import { selectSelectedGateway } from '@console/store/selectors/gateways'

const m = defineMessages({
  downloadGlobalConf: 'Download global_conf.json',
  globalConf: 'Global configuration',
  globalConfFailed: 'Failed to download global_conf.json',
  globalConfFailedMessage:
    'An unknown error occurred and the global_conf.json could not be downloaded',
  globalConfUnavailable: 'Unavailable for gateways without frequency plan',
})

const GatewayOverview = () => {
  const { gtwId } = useParams()
  const mayViewGatewayConf = useSelector(state => checkFromState(mayViewGatewayConfJson, state))
  const gateway = useSelector(selectSelectedGateway)
  const { ids, description, created_at, updated_at, frequency_plan_ids, gateway_server_address } =
    gateway

  const handleGlobalConfDownload = useCallback(async () => {
    try {
      const globalConf = await tts.Gateways.getGlobalConf(gtwId)
      const globalConfDataUri = composeDataUri(JSON.stringify(globalConf, undefined, 2))
      downloadDataUriAsFile(globalConfDataUri, 'global_conf.json')
    } catch (err) {
      toast({
        title: m.globalConfFailed,
        message: m.globalConfFailedMessage,
        type: toast.types.ERROR,
      })
    }
  }, [gtwId])

  const sheetData = [
    {
      header: sharedMessages.generalInformation,
      items: [
        {
          key: sharedMessages.gatewayID,
          value: gtwId,
          type: 'code',
          sensitive: false,
        },
        {
          key: sharedMessages.gatewayEUI,
          value: ids.eui,
          type: 'byte',
          sensitive: false,
        },
        {
          key: sharedMessages.gatewayDescription,
          value: description || <Message content={sharedMessages.none} />,
        },
        {
          key: sharedMessages.createdAt,
          value: <DateTime value={created_at} />,
        },
        {
          key: sharedMessages.updatedAt,
          value: <DateTime value={updated_at} />,
        },
        {
          key: sharedMessages.gatewayServerAddress,
          value: gateway_server_address,
          type: 'code',
          sensitive: false,
        },
      ],
    },
  ]

  const lorawanInfo = {
    header: sharedMessages.lorawanInformation,
    items: [
      {
        key: sharedMessages.frequencyPlan,
        value:
          frequency_plan_ids.length !== 0 ? (
            <Message content={frequency_plan_ids.join(' , ')} />
          ) : undefined,
      },
    ],
  }

  if (mayViewGatewayConf) {
    lorawanInfo.items.push({
      key: m.globalConf,
      value:
        frequency_plan_ids.length !== 0 ? (
          <Button
            type="button"
            icon={IconFileDownload}
            onClick={handleGlobalConfDownload}
            message={m.downloadGlobalConf}
            secondary
          />
        ) : (
          <Message content={m.globalConfUnavailable} className="c-text-neutral-light" />
        ),
    })
  }

  sheetData.push(lorawanInfo)

  return (
    <Require featureCheck={mayViewGatewayInfo} otherwise={{ redirect: '/' }}>
      <IntlHelmet title={sharedMessages.overview} />
      <GatewayOverviewHeader gateway={gateway} />
      <div className="container container--xl grid p-ls-s gap-ls-s">
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <div style={{ height: '30rem', backgroundColor: 'lightgray' }} />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <div style={{ height: '30rem', backgroundColor: 'lightgray' }} />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <div style={{ height: '30rem', backgroundColor: 'lightgray' }} />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <div style={{ height: '30rem', backgroundColor: 'lightgray' }} />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6" />
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <BlurryNetworkActivityPanel />
        </div>
      </div>
    </Require>
  )
}

export default GatewayOverview
