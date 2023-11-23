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

import React, { useMemo, useCallback, useRef } from 'react'
import { Container, Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'
import { useSelector } from 'react-redux'

import AppAnimation from '@assets/animations/illustrations/app.json'
import GatewayAnimation from '@assets/animations/illustrations/gateway.json'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import Animation from '@ttn-lw/lib/components/animation'
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
  const appAnimationRef = useRef(null)
  const gatewayAnimationRef = useRef(null)

  useBreadcrumbs('overview', <Breadcrumb path="/" content={sharedMessages.overview} />)

  const handleAppChooserMouseEnter = useCallback(() => {
    appAnimationRef.current.setDirection(1)
    appAnimationRef.current.goToAndPlay(0)
  }, [appAnimationRef])

  const handleAppChooserMouseLeave = useCallback(() => {
    appAnimationRef.current.setDirection(-1)
  }, [appAnimationRef])

  const handleGatewayChooserMouseEnter = useCallback(() => {
    gatewayAnimationRef.current.setDirection(1)
    gatewayAnimationRef.current.goToAndPlay(0)
  }, [gatewayAnimationRef])

  const handleGatewayChooserMouseLeave = useCallback(() => {
    gatewayAnimationRef.current.setDirection(-1)
  }, [gatewayAnimationRef])

  const chooser = useMemo(() => {
    const hasEntities = applicationCount + gatewayCount !== 0
    const appPath = hasEntities ? '/applications' : '/applications/add'
    const gatewayPath = hasEntities ? '/gateways' : '/gateways/add'

    return (
      <Row>
        {mayViewApplications && (
          <Col lg={mayViewGateways ? 6 : 12}>
            <Link to={appPath} className={style.chooserNav}>
              <div
                onMouseEnter={handleAppChooserMouseEnter}
                onMouseLeave={handleAppChooserMouseLeave}
                className={style.chooser}
              >
                <Animation animationRef={appAnimationRef} animationData={AppAnimation} />
                <Message
                  component="span"
                  content={hasEntities ? m.gotoApplications : m.createApplication}
                />
              </div>
            </Link>
          </Col>
        )}
        {mayViewGateways && (
          <Col lg={mayViewApplications ? 6 : 12}>
            <Link to={gatewayPath} className={style.chooserNav}>
              <div
                onMouseEnter={handleGatewayChooserMouseEnter}
                onMouseLeave={handleGatewayChooserMouseLeave}
                className={style.chooser}
              >
                <Animation animationRef={gatewayAnimationRef} animationData={GatewayAnimation} />
                <Message
                  component="span"
                  content={hasEntities ? m.gotoGateways : m.createGateway}
                />
              </div>
            </Link>
          </Col>
        )}
      </Row>
    )
  }, [
    appAnimationRef,
    applicationCount,
    gatewayAnimationRef,
    gatewayCount,
    handleAppChooserMouseEnter,
    handleAppChooserMouseLeave,
    handleGatewayChooserMouseEnter,
    handleGatewayChooserMouseLeave,
  ])

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
              <Message
                className={style.welcome}
                content={hasEntities ? m.welcomeBack : m.welcome}
                values={{ userName }}
                component="h1"
              />
              {!mayNotViewEntities && (
                <Message
                  className={style.getStarted}
                  content={hasEntities || !mayCreateEntities ? m.continueWorking : m.getStarted}
                  component="h2"
                />
              )}
              <HelpLink supportLink={supportLink} documentationLink={documentationBaseUrl} />
            </Col>
          </Row>
          {chooser}
        </div>
        <DeploymentComponentStatus />
      </Container>
    </RequireRequest>
  )
}

export default Overview
