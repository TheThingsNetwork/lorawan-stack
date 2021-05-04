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

import ServerIcon from '@assets/auxiliary-icons/server.svg'
import AppAnimation from '@assets/animations/illustrations/app.json'
import GatewayAnimation from '@assets/animations/illustrations/gateway.json'

import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Status from '@ttn-lw/components/status'
import Spinner from '@ttn-lw/components/spinner'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import { withEnv } from '@ttn-lw/lib/components/env'
import Animation from '@ttn-lw/lib/components/animation'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import { createFetchingSelector } from '@ttn-lw/lib/store/selectors/fetching'

import {
  mayViewApplications,
  mayViewGateways,
  mayCreateApplications,
  mayCreateGateways,
} from '@console/lib/feature-checks'

import { getApplicationsList, GET_APPS_LIST_BASE } from '@console/store/actions/applications'
import { getGatewaysList, GET_GTWS_LIST_BASE } from '@console/store/actions/gateways'

import { selectApplicationsTotalCount } from '@console/store/selectors/applications'
import { selectGatewaysTotalCount } from '@console/store/selectors/gateways'
import { selectUserNameOrId, selectUserRights } from '@console/store/selectors/user'

import style from './overview.styl'

const m = defineMessages({
  createApplication: 'Create an application',
  createGateway: 'Register a gateway',
  gotoApplications: 'Go to applications',
  gotoGateways: 'Go to gateways',
  needHelp: 'Need help? Have a look at our {documentationLink} or {supportLink}.',
  needHelpShort: 'Need help? Have a look at our {link}.',
  welcome: 'Welcome to the Console!',
  welcomeBack: 'Welcome back, {userName}! 👋',
  getStarted: 'Get started right away by creating an application or registering a gateway.',
  continueWorking: 'Walk right through to your applications and/or gateways.',
  componentStatus: 'Component status',
  versionInfo: 'Version info',
})

const componentMap = {
  is: sharedMessages.componentIs,
  gs: sharedMessages.componentGs,
  ns: sharedMessages.componentNs,
  as: sharedMessages.componentAs,
  js: sharedMessages.componentJs,
}

const overviewFetchingSelector = createFetchingSelector([GET_APPS_LIST_BASE, GET_GTWS_LIST_BASE])

@connect(
  state => {
    const rights = selectUserRights(state)

    return {
      applicationCount: selectApplicationsTotalCount(state),
      gatewayCount: selectGatewaysTotalCount(state),
      fetching: overviewFetchingSelector(state),
      userName: selectUserNameOrId(state),
      mayCreateApplications: mayCreateApplications.check(rights),
      mayViewApplications: mayViewApplications.check(rights),
      mayViewGateways: mayViewGateways.check(rights),
      mayCreateGateways: mayCreateGateways.check(rights),
    }
  },
  dispatch => ({
    loadData: () => {
      dispatch(getApplicationsList())
      dispatch(getGatewaysList())
    },
  }),
)
@withBreadcrumb('overview', props => <Breadcrumb path="/" content={sharedMessages.overview} />)
@withEnv
export default class Overview extends React.Component {
  static propTypes = {
    applicationCount: PropTypes.number,
    env: PropTypes.env,
    fetching: PropTypes.bool.isRequired,
    gatewayCount: PropTypes.number,
    loadData: PropTypes.func.isRequired,
    mayCreateApplications: PropTypes.bool.isRequired,
    mayCreateGateways: PropTypes.bool.isRequired,
    mayViewApplications: PropTypes.bool.isRequired,
    mayViewGateways: PropTypes.bool.isRequired,
    userName: PropTypes.string.isRequired,
  }

  static defaultProps = {
    applicationCount: 0,
    env: undefined,
    gatewayCount: 0,
  }

  constructor(props) {
    super(props)

    this.appAnimationRef = React.createRef()
    this.gatewayAnimationRef = React.createRef()
  }

  componentDidMount() {
    const { loadData } = this.props
    loadData()
  }

  @bind
  handleAppChooserMouseEnter() {
    this.appAnimationRef.current.instance.setDirection(1)
    this.appAnimationRef.current.instance.goToAndPlay(0)
  }

  @bind
  handleAppChooserMouseLeave() {
    this.appAnimationRef.current.instance.setDirection(-1)
  }

  @bind
  handleGatewayChooserMouseEnter() {
    this.gatewayAnimationRef.current.instance.setDirection(1)
    this.gatewayAnimationRef.current.instance.goToAndPlay(0)
  }

  @bind
  handleGatewayChooserMouseLeave() {
    this.gatewayAnimationRef.current.instance.setDirection(-1)
  }

  get chooser() {
    const { applicationCount, gatewayCount, mayViewApplications, mayViewGateways } = this.props
    const hasEntities = applicationCount + gatewayCount !== 0
    const appPath = hasEntities ? '/applications' : '/applications/add'
    const gatewayPath = hasEntities ? '/gateways' : '/gateways/add'

    return (
      <Row>
        {mayViewApplications && (
          <Col lg={mayViewGateways ? 6 : 12}>
            <Link to={appPath} className={style.chooserNav}>
              <div
                onMouseEnter={this.handleAppChooserMouseEnter}
                onMouseLeave={this.handleAppChooserMouseLeave}
                className={style.chooser}
              >
                <Animation ref={this.appAnimationRef} animationData={AppAnimation} />
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
                onMouseEnter={this.handleGatewayChooserMouseEnter}
                onMouseLeave={this.handleGatewayChooserMouseLeave}
                className={style.chooser}
              >
                <Animation ref={this.gatewayAnimationRef} animationData={GatewayAnimation} />
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
  }

  render() {
    const {
      config: { stack: stackConfig, supportLink, documentationBaseUrl },
    } = this.props.env
    const {
      fetching,
      applicationCount,
      gatewayCount,
      userName,
      mayCreateApplications,
      mayCreateGateways,
      mayViewApplications,
      mayViewGateways,
    } = this.props
    const hasEntities = applicationCount + gatewayCount !== 0
    const mayCreateEntities = mayCreateApplications || mayCreateGateways
    const mayNotViewEntities = !mayViewApplications && !mayViewGateways

    if (fetching || applicationCount === undefined || gatewayCount === undefined) {
      return (
        <Spinner center>
          <Message content={sharedMessages.fetching} />
        </Spinner>
      )
    }

    return (
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
          {this.chooser}
        </div>
        <hr />
        <Row className={style.componentSection}>
          <Col sm={4} className={style.versionInfoSection}>
            <Message content={m.versionInfo} component="h3" />
            <span className={style.versionValue}>v{process.env.VERSION}</span>
          </Col>
          <Col sm={8}>
            <Message className={style.componentStatus} content={m.componentStatus} component="h3" />
            <div className={style.componentCards}>
              {Object.keys(stackConfig).map(componentKey => {
                if (componentKey === 'language') {
                  return null
                }
                const component = stackConfig[componentKey]
                const name = componentMap[componentKey]
                const host = component.enabled ? new URL(component.base_url).host : undefined
                return (
                  <ComponentCard
                    key={componentKey}
                    name={name}
                    host={host}
                    enabled={component.enabled}
                  />
                )
              })}
            </div>
          </Col>
        </Row>
      </Container>
    )
  }
}

const ComponentCard = ({ name, enabled, host }) => (
  <div className={style.componentCard}>
    <img src={ServerIcon} className={style.componentCardIcon} />
    <div className={style.componentCardDesc}>
      <div className={style.componentCardName}>
        <Status label={name} status={enabled ? 'good' : 'unknown'} flipped />
      </div>
      <span className={style.componentCardHost} title={host}>
        {enabled ? host : <Message content={sharedMessages.disabled} />}
      </span>
    </div>
  </div>
)

ComponentCard.propTypes = {
  enabled: PropTypes.bool.isRequired,
  host: PropTypes.string,
  name: PropTypes.message.isRequired,
}

ComponentCard.defaultProps = {
  host: undefined,
}

const HelpLink = ({ supportLink, documentationLink }) => {
  if (!supportLink && !documentationLink) return null

  const documentation = (
    <Link.DocLink secondary path="/" title={sharedMessages.documentation}>
      <Message content={sharedMessages.documentation} />
    </Link.DocLink>
  )

  const support = (
    <Link.Anchor secondary href={supportLink || ''} external>
      <Message content={sharedMessages.getSupport} />
    </Link.Anchor>
  )

  return (
    <Message
      className={style.getStarted}
      content={documentationLink && supportLink ? m.needHelp : m.needHelpShort}
      values={{
        documentationLink: documentation,
        supportLink: support,
        link: documentationLink ? documentation : support,
      }}
      component="h2"
    />
  )
}

HelpLink.propTypes = {
  documentationLink: PropTypes.string,
  supportLink: PropTypes.string,
}

HelpLink.defaultProps = {
  supportLink: undefined,
  documentationLink: undefined,
}
