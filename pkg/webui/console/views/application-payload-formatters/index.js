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
import { Switch, Route, Redirect } from 'react-router-dom'
import { Container, Col, Row } from 'react-grid-system'

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
  selectApplicationLink,
  selectApplicationLinkFetching,
  selectSelectedApplicationId,
} from '@console/store/selectors/applications'

@connect(
  state => {
    const link = selectApplicationLink(state)
    const fetching = selectApplicationLinkFetching(state)
    return {
      appId: selectSelectedApplicationId(state),
      fetching: fetching || !link,
    }
  },
  dispatch => ({
    getLink: (id, selector) => dispatch(getApplicationLink(id, selector)),
  }),
)
@withBreadcrumb('apps.single.payload-formatters', props => {
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
    match: PropTypes.match.isRequired,
  }

  componentDidMount() {
    const { appId, getLink } = this.props

    getLink(appId, ['default_formatters'])
  }

  render() {
    const {
      match,
      fetching,
      match: { url },
    } = this.props

    if (fetching) {
      return (
        <Spinner center>
          <Message content={sharedMessages.fetching} />
        </Spinner>
      )
    }

    return (
      <Container>
        <Row>
          <Col>
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
