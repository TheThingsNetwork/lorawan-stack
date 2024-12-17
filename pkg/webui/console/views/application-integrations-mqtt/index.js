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

import React, { useCallback, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'
import { useParams } from 'react-router-dom'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import DataSheet from '@ttn-lw/components/data-sheet'
import Button from '@ttn-lw/components/button'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'
import ErrorView from '@ttn-lw/lib/components/error-view'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import Require from '@console/lib/components/require'

import SubViewError from '@console/views/sub-view-error'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayViewMqttConnectionInfo } from '@console/lib/feature-checks'

import { createApplicationApiKey } from '@console/store/actions/api-keys'
import { getMqttInfo } from '@console/store/actions/applications'

import { selectMqttConnectionInfo } from '@console/store/selectors/applications'

const m = defineMessages({
  publicAddress: 'Public address',
  publicTlsAddress: 'Public TLS address',
  generateApiKey: 'Generate new API key',
  goToApiKeys: 'Go to API keys',
  mqttInfoText:
    'MQTT is a publish/subscribe messaging protocol designed for IoT. Every application on TTS automatically exposes an MQTT endpoint. In order to connect to the MQTT server you need to create a new API key, which will function as connection password. You can also use an existing API key, as long as it has the necessary rights granted.',
  connectionCredentials: 'Connection credentials',
  mqttIntegrations: 'MQTT integrations',
  officialMqttWebsite: 'Official MQTT website',
  mqttServer: 'MQTT server',
  host: 'MQTT server host',
  connectionInfo: 'Connection information',
})

const ApplicationMqtt = () => {
  const { appId } = useParams()
  const connectionInfo = useSelector(selectMqttConnectionInfo)
  const [apiKey, setApiKey] = useState()
  const dispatch = useDispatch()

  useBreadcrumbs(
    'apps.single.integrations.mqtt',
    <Breadcrumb path={`/applications/${appId}/integrations/mqtt`} content={sharedMessages.mqtt} />,
  )

  const handleGeneratePasswordClick = useCallback(async () => {
    const key = {
      name: `mqtt-password-key-${Date.now()}`,
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ', 'RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE'],
    }
    const result = await dispatch(attachPromise(createApplicationApiKey(appId, key)))
    setApiKey(result)
  }, [appId, dispatch])

  const connectionData = [
    { header: m.host, items: [] },
    { header: m.connectionCredentials, items: [] },
  ]
  const fetchingMessage = <Message content={sharedMessages.fetching} />

  if (connectionInfo) {
    const { public_address, public_tls_address, username } = connectionInfo
    connectionData[0].items = [
      {
        key: m.publicAddress,
        type: 'code',
        sensitive: false,
        value: public_address,
      },
      {
        key: m.publicTlsAddress,
        type: 'code',
        sensitive: false,
        value: public_tls_address,
      },
    ]
    connectionData[1].items = [
      {
        key: sharedMessages.username,
        type: 'code',
        sensitive: false,
        value: username,
      },
    ]
  } else {
    connectionData[0].items = [
      {
        key: m.publicAddress,
        value: fetchingMessage,
      },
      {
        key: m.publicTlsAddress,
        value: fetchingMessage,
      },
    ]
    connectionData[1].items = [
      {
        key: sharedMessages.username,
        value: fetchingMessage,
      },
    ]
  }
  if (apiKey) {
    connectionData[1].items.push({
      key: sharedMessages.password,
      type: 'code',
      value: apiKey.key,
    })
  } else {
    connectionData[1].items.push({
      key: sharedMessages.password,
      value: (
        <>
          <Button
            message={m.generateApiKey}
            onClick={handleGeneratePasswordClick}
            className="mr-cs-s"
            secondary
          />
          <Link to={`/applications/${appId}/api-keys`} naked secondary>
            <Message content={m.goToApiKeys} />
          </Link>
        </>
      ),
    })
  }

  return (
    <RequireRequest requestAction={getMqttInfo(appId)}>
      <Require
        featureCheck={mayViewMqttConnectionInfo}
        otherwise={{ redirect: `/applications/${appId}` }}
      >
        <ErrorView errorRender={SubViewError}>
          <div className="container container--xl grid">
            <div className="item-12">
              <PageTitle title={sharedMessages.mqtt} noGrid />
              <Message content={m.mqttInfoText} className="mt-0" />
              <div>
                <Message
                  component="h4"
                  content={sharedMessages.furtherResources}
                  className="mb-cs-xxs"
                />
                <Link.DocLink path="/integrations/mqtt" secondary>
                  <Message content={m.mqttServer} />
                </Link.DocLink>
                {' | '}
                <Link.Anchor href="https://www.mqtt.org" external secondary>
                  <Message content={m.officialMqttWebsite} />
                </Link.Anchor>
              </div>
              <hr className="mb-ls-s" />
              <Message content={m.connectionInfo} component="h3" />
              <DataSheet data={connectionData} />
            </div>
          </div>
        </ErrorView>
      </Require>
    </RequireRequest>
  )
}

export default ApplicationMqtt
