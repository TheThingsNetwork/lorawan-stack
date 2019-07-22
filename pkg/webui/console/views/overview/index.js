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
import { defineMessages } from 'react-intl'
import bind from 'autobind-decorator'
import { connect } from 'react-redux'

import sharedMessages from '../../../lib/shared-messages'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import { withEnv } from '../../../lib/components/env'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import IntlHelmet from '../../../lib/components/intl-helmet'
import Message from '../../../lib/components/message'
import Status from '../../../components/status'
import Spinner from '../../../components/spinner'
import Link from '../../../components/link'
import Animation from '../../../lib/components/animation'
import { selectApplicationsTotalCount } from '../../store/selectors/applications'
import { getApplicationsList, GET_APPS_LIST } from '../../store/actions/applications'
import { selectGatewaysTotalCount } from '../../store/selectors/gateways'
import { getGatewaysList, GET_GTWS_LIST } from '../../store/actions/gateways'
import { createFetchingSelector } from '../../store/selectors/fetching'
import { selectUserId } from '../../store/selectors/user'

import ServerIcon from '../../../assets/auxiliary-icons/server.svg'
import AppAnimation from '../../../assets/animations/illustrations/app.json'
import GatewayAnimation from '../../../assets/animations/illustrations/gateway.json'

import style from './overview.styl'

const m = defineMessages({
  createApplication: 'Create an application',
  createGateway: 'Register a gateway',
  gotoApplications: 'Go to applications',
  gotoGateways: 'Go to gateways',
  welcome: 'Welcome to the Console!',
  welcomeBack: 'Welcome back, {userId} 👋!',
  getStarted: 'Get started right away by creating an application or registering a gateway.',
  continueWorking: 'Walk right through to your applications and/or gateways.',
  componentStatus: 'Component Status',
  versionInfo: 'Version Info',
})

const componentMap = {
  is: sharedMessages.componentIdentityServer,
  gs: sharedMessages.componentGatewayServer,
  ns: sharedMessages.componentNetworkServer,
  as: sharedMessages.componentApplicationServer,
  js: sharedMessages.componentJoinServer,
}

@connect(state => ({
  applicationCount: selectApplicationsTotalCount(state),
  gatewayCount: selectGatewaysTotalCount(state),
  fetching: createFetchingSelector([ GET_APPS_LIST, GET_GTWS_LIST ])(state),
  userId: selectUserId(state),
}), dispatch => ({
  loadData () {
    dispatch(getApplicationsList())
    dispatch(getGatewaysList())
  },
}))
@withBreadcrumb('overview', function (props) {
  return (
    <Breadcrumb
      path="/console"
      content={sharedMessages.overview}
    />
  )
})
@withEnv
@bind
export default class Overview extends React.Component {

  constructor (props) {
    super(props)

    this.appAnimationRef = React.createRef()
    this.gatewayAnimationRef = React.createRef()
  }

  componentDidMount () {
    const { loadData } = this.props
    loadData()
  }

  handleAppChooserMouseEnter () {
    this.appAnimationRef.current.instance.setDirection(1)
    this.appAnimationRef.current.instance.goToAndPlay(0)
  }

  handleAppChooserMouseLeave () {
    this.appAnimationRef.current.instance.setDirection(-1)
  }

  handleGatewayChooserMouseEnter () {
    this.gatewayAnimationRef.current.instance.setDirection(1)
    this.gatewayAnimationRef.current.instance.goToAndPlay(0)
  }

  handleGatewayChooserMouseLeave () {
    this.gatewayAnimationRef.current.instance.setDirection(-1)
  }

  render () {
    const { config } = this.props.env
    const { fetching, applicationCount, gatewayCount, userId } = this.props
    const { path } = this.props.match

    if (fetching || applicationCount === undefined || gatewayCount === undefined) {
      return (
        <Spinner center>
          <Message content={sharedMessages.loading} />
        </Spinner>
      )
    }

    const hasEntities = applicationCount + gatewayCount !== 0
    const appPath = path + (hasEntities ? '/applications' : '/applications/add')
    const gatewayPath = path + (hasEntities ? '/gateways' : '/gateways/add')

    return (
      <Container>
        <Row className={style.welcomeSection}>
          <IntlHelmet title={sharedMessages.overview} />
          <Col sm={12} className={style.welcomeTitleSection}>
            <Message className={style.welcome} content={hasEntities ? m.welcomeBack : m.welcome} values={{ userId }} component="h1" />
            <Message className={style.getStarted} content={hasEntities ? m.continueWorking : m.getStarted} component="h2" />
          </Col>
          <Col lg={6}>
            <Link to={appPath} className={style.chooserNav}>
              <div
                onMouseEnter={this.handleAppChooserMouseEnter}
                onMouseLeave={this.handleAppChooserMouseLeave}
                className={style.chooser}
              >
                <Animation ref={this.appAnimationRef} animationData={AppAnimation} />
                <Message content={hasEntities ? m.gotoApplications : m.createApplication} />
              </div>
            </Link>
          </Col>
          <Col lg={6}>
            <Link to={gatewayPath} className={style.chooserNav}>
              <div
                onMouseEnter={this.handleGatewayChooserMouseEnter}
                onMouseLeave={this.handleGatewayChooserMouseLeave}
                className={style.chooser}
              >
                <Animation ref={this.gatewayAnimationRef} animationData={GatewayAnimation} />
                <Message content={hasEntities ? m.gotoGateways : m.createGateway} />
              </div>
            </Link>
          </Col>
        </Row>
        <hr />
        <Row className={style.componentSection}>
          <Col sm={4} className={style.versionInfoSection}>
            <Message content={m.versionInfo} component="h3" />
            <span className={style.versionValue}>v{process.env.VERSION}</span>
          </Col>
          <Col sm={8}>
            <Message className={style.componentStatus} content={m.componentStatus} component="h3" />
            <div className={style.componentCards}>
              { Object.keys(config).map(function (componentKey) {
                if (componentKey === 'language') {
                  return null
                }
                const component = config[componentKey]
                const name = componentMap[componentKey]
                const host = new URL(component.base_url).host
                return <ComponentCard key={componentKey} name={name} host={host} enabled={component.enabled} />
              })}
            </div>
          </Col>
        </Row>
      </Container>
    )
  }
}

const ComponentCard = function ({ name, enabled, host }) {

  return (
    <div className={style.componentCard}>
      <img src={ServerIcon} className={style.componentCardIcon} />
      <div className={style.componentCardDesc}>
        <div className={style.componentCardName}>
          <Status status={enabled ? 'good' : 'bad'} /><Message content={name} />
        </div>
        <span className={style.componentCardHost} title={host}>{host}</span>
      </div>
    </div>
  )
}
