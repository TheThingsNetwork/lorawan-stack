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

import api from '@console/api'

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

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayViewMqttConnectionInfo } from '@console/lib/feature-checks'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'

import style from './application-integrations-mqtt.styl'

const m = defineMessages({
  publicAddress: 'Public address',
  publicTlsAddress: 'Public TLS address',
  generateApiKey: 'Generate new API key',
  goToApiKeys: 'Go to API keys',
  mqttInfoText:
    'The Application Server exposes an MQTT server to work with streaming events. In order to use the MQTT server you need to create a new API key, which will function as connection password. You can also use an existing API key, as long as it has the necessary rights granted. Use the connection information below to connect.',
  connectionCredentials: 'Connection credentials',
  mqttIntegrations: 'MQTT integrations',
})

@connect(state => ({
  appId: selectSelectedApplicationId(state),
}))
@withFeatureRequirement(mayViewMqttConnectionInfo, {
  redirect: ({ appId }) => `/applications/${appId}`,
})
@withBreadcrumb('apps.single.integrations.mqtt', function (props) {
  const { appId } = props

  return (
    <Breadcrumb path={`/applications/${appId}/integrations/mqtt`} content={sharedMessages.mqtt} />
  )
})
export default class ApplicationMqtt extends React.Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
  }

  state = {
    connectionInfo: undefined,
  }

  async componentDidMount() {
    const { appId } = this.props
    const connectionInfo = await api.application.getMqttConnectionInfo(appId)

    this.setState({ connectionInfo })
  }

  @bind
  async handleGeneratePasswordClick() {
    const { appId } = this.props
    const key = {
      name: `mqtt-password-key-${Date.now()}`,
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ', 'RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE'],
    }
    const result = await api.application.apiKeys.create(appId, key)

    this.setState({
      key: result,
    })
  }

  render() {
    const { appId } = this.props
    const { connectionInfo, key } = this.state
    const connectionData = [{ header: m.connectionCredentials, items: [] }]
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
        {
          key: sharedMessages.username,
          value: fetchingMessage,
        },
      ]
    }

    if (key) {
      connectionData[0].items.push({
        key: sharedMessages.password,
        type: 'code',
        value: key.key,
      })
    } else {
      connectionData[0].items.push({
        key: sharedMessages.password,
        value: (
          <React.Fragment>
            <Button
              className={style.generateApiKeyButton}
              message={m.generateApiKey}
              onClick={this.handleGeneratePasswordClick}
            />
            <Link to={`/applications/${appId}/api-keys`} secondary>
              <Message content={m.goToApiKeys} />
            </Link>
          </React.Fragment>
        ),
      })
    }

    return (
      <ErrorView ErrorComponent={SubViewError}>
        <Container>
          <PageTitle title={sharedMessages.mqtt} />
          <Row>
            <Col lg={8} md={12}>
              <Message component="p" content={m.mqttInfoText} className={style.info} />
              <hr className={style.hRule} />
              <DataSheet data={connectionData} />
            </Col>
          </Row>
        </Container>
      </ErrorView>
    )
  }
}
