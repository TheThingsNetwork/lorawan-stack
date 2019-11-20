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
import { connect } from 'react-redux'
import { Switch, Route, Link } from 'react-router-dom'
import { Container, Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import PropTypes from '../../../lib/prop-types'
import Message from '../../../lib/components/message'
import Notification from '../../../components/notification'
import Spinner from '../../../components/spinner'
import sharedMessages from '../../../lib/shared-messages'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import NotFoundRoute from '../../../lib/components/not-found-route'

import ApplicationUplinkPayloadFormatters from '../../containers/application-payload-formatters/uplink'
import ApplicationDownlinkPayloadFormatters from '../../containers/application-payload-formatters/downlink'

import { getApplicationLink } from '../../store/actions/link'
import {
  selectApplicationIsLinked,
  selectApplicationLink,
  selectApplicationLinkFetching,
  selectSelectedApplicationId,
} from '../../store/selectors/applications'

import style from './application-payload-formatters.styl'

const m = defineMessages({
  warningTitle: 'Linking Needed',
  warningText: 'Please {link}, in order to configure payload formatters',
  linkApplication: 'link your application',
})

@connect(
  function(state) {
    const link = selectApplicationLink(state)
    const fetching = selectApplicationLinkFetching(state)

    return {
      appId: selectSelectedApplicationId(state),
      fetching: fetching || !link,
      linked: selectApplicationIsLinked(state),
    }
  },
  dispatch => ({
    getLink: (id, selector) => dispatch(getApplicationLink(id, selector)),
  }),
)
@withBreadcrumb('apps.single.payload-formatters', function(props) {
  const { appId } = props

  return (
    <Breadcrumb
      path={`/applications/${appId}/payload-formatters`}
      icon="link"
      content={sharedMessages.payloadFormatters}
    />
  )
})
export default class ApplicationPayloadFormatters extends React.Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    fetching: PropTypes.bool.isRequired,
    getLink: PropTypes.func.isRequired,
    linked: PropTypes.bool.isRequired,
    match: PropTypes.match.isRequired,
  }

  componentDidMount() {
    const { appId, getLink } = this.props

    getLink(appId, ['default_formatters', 'api_key', 'network_server_address'])
  }

  render() {
    const { match, fetching, linked, appId } = this.props

    if (fetching) {
      return (
        <Spinner center>
          <Message content={sharedMessages.fetching} />
        </Spinner>
      )
    }

    const linkWarning = linked ? null : (
      <Notification
        className={style.warningNotification}
        title={m.warningTitle}
        warning
        content={m.warningText}
        messageValues={{
          link: (
            <Link key="warnining-link" to={`/applications/${appId}/link`}>
              <Message content={m.linkApplication} />
            </Link>
          ),
        }}
      />
    )

    return (
      <Container>
        <Row>
          <Col>
            {linkWarning}
            <Switch>
              <Route path={`${match.url}/uplink`} component={ApplicationUplinkPayloadFormatters} />
              <Route
                path={`${match.url}/downlink`}
                component={ApplicationDownlinkPayloadFormatters}
              />
              <NotFoundRoute />
            </Switch>
          </Col>
        </Row>
      </Container>
    )
  }
}
