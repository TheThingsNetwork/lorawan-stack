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
import { Container, Col, Row } from 'react-grid-system'
import { connect } from 'react-redux'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import DataSheet from '@ttn-lw/components/data-sheet'
import Button from '@ttn-lw/components/button'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'
import ErrorView from '@ttn-lw/lib/components/error-view'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import SubViewError from '@console/views/sub-view-error'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayViewMqttConnectionInfo } from '@console/lib/feature-checks'

import { createApplicationApiKey } from '@console/store/actions/api-keys'
import { getMqttInfo } from '@console/store/actions/applications'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'

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

@connect(
  state => ({
    appId: selectSelectedApplicationId(state),
  }),
  {
    createApiKey: (appId, key) => attachPromise(createApplicationApiKey(appId, key)),
    getMqttConnectionInfo: id => attachPromise(getMqttInfo(id)),
  },
)
@withFeatureRequirement(mayViewMqttConnectionInfo, {
  redirect: ({ appId }) => `/applications/${appId}`,
})
@withBreadcrumb('apps.single.integrations.mqtt', props => {
  const { appId } = props

  return (
    <Breadcrumb path={`/applications/${appId}/integrations/mqtt`} content={sharedMessages.mqtt} />
  )
})
export default class ApplicationMqtt extends React.Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    createApiKey: PropTypes.func.isRequired,
    getMqttConnectionInfo: PropTypes.func.isRequired,
  }

  state = {
    connectionInfo: undefined,
  }

  async componentDidMount() {
    const { appId, getMqttConnectionInfo } = this.props
    const connectionInfo = await getMqttConnectionInfo(appId)

    this.setState({ connectionInfo })
  }

  @bind
  async handleGeneratePasswordClick() {
    const { appId, createApiKey } = this.props
    const key = {
      name: `mqtt-password-key-${Date.now()}`,
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ', 'RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE'],
    }
    const result = await createApiKey(appId, key)

    this.setState({
      key: result,
    })
  }

  render() {
    const { appId } = this.props
    const { connectionInfo, key } = this.state
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

    if (key) {
      connectionData[1].items.push({
        key: sharedMessages.password,
        type: 'code',
        value: key.key,
      })
    } else {
      connectionData[1].items.push({
        key: sharedMessages.password,
        value: (
          <>
            <Button
              message={m.generateApiKey}
              onClick={this.handleGeneratePasswordClick}
              className="mr-cs-s"
            />
            <Link to={`/applications/${appId}/api-keys`} naked secondary>
              <Message content={m.goToApiKeys} />
            </Link>
          </>
        ),
      })
    }

    return (
      <ErrorView errorRender={SubViewError}>
        <Container>
          <PageTitle title={sharedMessages.mqtt} />
          <Row>
            <Col lg={8} md={12}>
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
            </Col>
          </Row>
        </Container>
      </ErrorView>
    )
  }
}
