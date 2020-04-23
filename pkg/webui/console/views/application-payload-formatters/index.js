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
import { Switch, Route, Link, Redirect } from 'react-router-dom'
import { Container, Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import Notification from '@ttn-lw/components/notification'
import Spinner from '@ttn-lw/components/spinner'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'
import Message from '@ttn-lw/lib/components/message'

import ApplicationUplinkPayloadFormatters from '@console/containers/application-payload-formatters/uplink'
import ApplicationDownlinkPayloadFormatters from '@console/containers/application-payload-formatters/downlink'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { getApplicationLink } from '@console/store/actions/link'

import {
  selectApplicationIsLinked,
  selectApplicationLink,
  selectApplicationLinkFetching,
  selectSelectedApplicationId,
} from '@console/store/selectors/applications'

import style from './application-payload-formatters.styl'

const m = defineMessages({
  warningTitle: 'Linking needed',
  warningText: 'Please {link}, in order to configure payload formatters',
  linkApplication: 'Link your application',
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
    const {
      match,
      fetching,
      linked,
      appId,
      match: { url },
    } = this.props

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
              <Redirect exact from={url} to={`${url}/uplink`} />
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
