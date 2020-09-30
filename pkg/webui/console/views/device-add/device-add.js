// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import { Container, Col, Row } from 'react-grid-system'
import { Switch, Route } from 'react-router'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { getDeviceId } from '@ttn-lw/lib/selectors/id'
import PropTypes from '@ttn-lw/lib/prop-types'

import ConfigurationForm from './configuration-form'
import DeviceWizard from './wizard'

const m = defineMessages({
  title: 'Add new end device',
})

const DeviceAdd = React.memo(props => {
  const {
    match,
    location,
    appId,
    redirectToWizard,
    redirectToEndDevice,
    redirectToConfiguration,
    redirectToNextLocation,
    jsConfig,
    nsConfig,
    asConfig,
    mayEditKeys,
    createDevice,
    prefixes,
  } = props

  const [configuration, setConfiguration] = React.useState(undefined)
  const handleConfigurationSubmit = React.useCallback(
    configuration => {
      setConfiguration(configuration)

      redirectToWizard()
    },
    [redirectToWizard],
  )
  const hasConfiguration = typeof configuration !== 'undefined'

  const rollbackProgress = React.useCallback(
    nextLocation => {
      redirectToNextLocation(nextLocation)
    },
    [redirectToNextLocation],
  )

  const handleCreateDevice = React.useCallback(
    device => {
      return createDevice(appId, device)
    },
    [appId, createDevice],
  )

  const handleCreateDeviceSuccess = React.useCallback(
    device => {
      const deviceId = getDeviceId(device)

      redirectToEndDevice(appId, deviceId)
    },
    [appId, redirectToEndDevice],
  )

  React.useEffect(() => {
    if (location.pathname.endsWith('/steps') && !hasConfiguration) {
      redirectToConfiguration()
    }
  }, [hasConfiguration, location, redirectToConfiguration])

  return (
    <Container>
      <Row>
        <Col lg={8} md={12}>
          <PageTitle title={m.title} />
          <Switch>
            <Route exact path={`${match.url}`}>
              {() => (
                <ConfigurationForm
                  asConfig={asConfig}
                  jsConfig={jsConfig}
                  nsConfig={nsConfig}
                  onSubmit={handleConfigurationSubmit}
                  initialValues={configuration}
                  mayEditKeys={mayEditKeys}
                />
              )}
            </Route>
            <Route path={`${match.url}/steps`}>
              {({ match }) =>
                hasConfiguration ? (
                  <DeviceWizard
                    prefixes={prefixes}
                    createDevice={handleCreateDevice}
                    createDeviceSuccess={handleCreateDeviceSuccess}
                    rollbackProgress={rollbackProgress}
                    asConfig={asConfig}
                    jsConfig={jsConfig}
                    nsConfig={nsConfig}
                    mayEditKeys={mayEditKeys}
                    configuration={configuration}
                    match={match}
                  />
                ) : null
              }
            </Route>
          </Switch>
        </Col>
      </Row>
    </Container>
  )
})

DeviceAdd.propTypes = {
  appId: PropTypes.string.isRequired,
  asConfig: PropTypes.stackComponent.isRequired,
  createDevice: PropTypes.func.isRequired,
  jsConfig: PropTypes.stackComponent.isRequired,
  location: PropTypes.location.isRequired,
  match: PropTypes.match.isRequired,
  mayEditKeys: PropTypes.bool.isRequired,
  nsConfig: PropTypes.stackComponent.isRequired,
  prefixes: PropTypes.euiPrefixes,
  redirectToConfiguration: PropTypes.func.isRequired,
  redirectToEndDevice: PropTypes.func.isRequired,
  redirectToNextLocation: PropTypes.func.isRequired,
  redirectToWizard: PropTypes.func.isRequired,
}

DeviceAdd.defaultProps = {
  prefixes: [],
}

export default withBreadcrumb('devices.add', props => (
  <Breadcrumb path={`/applications/${props.appId}/devices/add`} content={sharedMessages.add} />
))(DeviceAdd)
