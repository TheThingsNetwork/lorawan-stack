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

import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import ErrorView from '../../../lib/components/error-view'
import SubViewError from '../error/sub-view'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import DataSheet from '../../../components/data-sheet'
import Button from '../../../components/button'
import api from '../../api'
import withFeatureRequirement from '../../lib/components/with-feature-requirement'

import { selectSelectedApplicationId } from '../../store/selectors/applications'
import { mayViewMqttConnectionInfo } from '../../lib/feature-checks'

import style from './application-integrations-mqtt.styl'

const m = defineMessages({
  publicAddress: 'Public Address',
  publicTlsAddress: 'Public TLS Address',
  generateApiKey: 'Generate new API Key',
  viewApiKeys: 'View API Keys',
  mqttInfoText:
    'The Application Server exposes an MQTT server to work with streaming events. In order to use the MQTT server you need to create a new API Key, which will function as connection password. You can also use an existing API Key, as long as it has the necessary rights granted. Use the connection information below to connect.',
  connectionCredentials: 'Connection Credentials',
  mqttIntegrations: 'MQTT Integrations',
})

@connect(state => ({
  appId: selectSelectedApplicationId(state),
}))
@withFeatureRequirement(mayViewMqttConnectionInfo, {
  redirect: ({ appId }) => `/applications/${appId}`,
})
@withBreadcrumb('apps.single.integrations.mqtt', function(props) {
  const { appId } = props

  return (
    <Breadcrumb
      path={`/applications/${appId}/integrations/mqtt`}
      icon="extension"
      content={sharedMessages.mqtt}
    />
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
            <Button message={m.generateApiKey} onClick={this.handleGeneratePasswordClick} />
            <Button.Link to={`/applications/${appId}/api-keys`} message={m.viewApiKeys} secondary />
          </React.Fragment>
        ),
      })
    }

    return (
      <ErrorView ErrorComponent={SubViewError}>
        <Container>
          <Row>
            <Col lg={8} md={12}>
              <IntlHelmet title={sharedMessages.mqtt} />
              <Message component="h2" content={m.mqttIntegrations} />
              <Message component="p" content={m.mqttInfoText} />
              <hr className={style.hRule} />
              <DataSheet data={connectionData} />
            </Col>
          </Row>
        </Container>
      </ErrorView>
    )
  }
}
