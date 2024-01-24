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
import { Container, Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'
import { useSelector } from 'react-redux'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import Message from '@ttn-lw/lib/components/message'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import DeploymentComponentStatus from '@console/containers/deployment-component-status'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectDocumentationUrlConfig, selectSupportLinkConfig } from '@ttn-lw/lib/selectors/env'

import {
  mayViewApplications,
  mayViewGateways,
  mayCreateApplications,
  mayCreateGateways,
} from '@console/lib/feature-checks'
import { checkFromState } from '@account/lib/feature-checks'

import { getApplicationsList } from '@console/store/actions/applications'
import { getGatewaysList } from '@console/store/actions/gateways'

import { selectApplicationsTotalCount } from '@console/store/selectors/applications'
import { selectGatewaysTotalCount } from '@console/store/selectors/gateways'
import { selectUserNameOrId } from '@console/store/selectors/logout'

import HelpLink from './help-link'

import style from './overview.styl'
import NotificationsDashboardPanel from '@console/containers/notifications-dashboard-panel'

const m = defineMessages({
  createApplication: 'Create an application',
  createGateway: 'Register a gateway',
  gotoApplications: 'Go to applications',
  gotoGateways: 'Go to gateways',
  welcome: 'Welcome to the Console!',
  welcomeBack: 'Welcome back, {userName}! 👋',
  getStarted: 'Get started right away by creating an application or registering a gateway.',
  continueWorking: 'Walk right through to your applications and/or gateways.',
  componentStatus: 'Component status',
  versionInfo: 'Version info',
})

const Overview = () => {
  const applicationCount = useSelector(selectApplicationsTotalCount)
  const gatewayCount = useSelector(selectGatewaysTotalCount)
  const userName = useSelector(selectUserNameOrId)
  const mayCreateApps = useSelector(state => checkFromState(mayCreateApplications, state))
  const mayViewApps = useSelector(state => checkFromState(mayViewApplications, state))
  const mayViewGtws = useSelector(state => checkFromState(mayViewGateways, state))
  const mayCreateGtws = useSelector(state => checkFromState(mayCreateGateways, state))
  const supportLink = selectSupportLinkConfig()
  const documentationBaseUrl = selectDocumentationUrlConfig()

  useBreadcrumbs('overview', <Breadcrumb path="/" content={sharedMessages.overview} />)

  const hasEntities = applicationCount + gatewayCount !== 0
  const mayCreateEntities = mayCreateApps || mayCreateGtws
  const mayNotViewEntities = !mayViewApps && !mayViewGtws

  return (
    <RequireRequest requestAction={[getApplicationsList(), getGatewaysList()]}>
      <Container>
        <div className={style.welcomeSection}>
          <Row>
            <IntlHelmet title={sharedMessages.overview} />
            <Col sm={12} className={style.welcomeTitleSection}>
              <NotificationsDashboardPanel />
            </Col>
          </Row>
        </div>
        <DeploymentComponentStatus />
      </Container>
    </RequireRequest>
  )
}

export default Overview
