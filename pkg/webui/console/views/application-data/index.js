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
import { defineMessages } from 'react-intl'
import { Container, Col, Row } from 'react-grid-system'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'

import ApplicationEvents from '@console/containers/application-events'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { mayViewApplicationEvents } from '@console/lib/feature-checks'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'

import style from './application-data.styl'

const m = defineMessages({
  appData: 'Application data',
})

@connect(state => ({ appId: selectSelectedApplicationId(state) }))
@withFeatureRequirement(mayViewApplicationEvents, {
  redirect: ({ appId }) => `/applications/${appId}`,
})
@withBreadcrumb('apps.single.data', function(props) {
  return <Breadcrumb path={`/applications/${props.appId}/data`} content={sharedMessages.data} />
})
export default class Data extends React.Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
  }

  render() {
    const { appId } = this.props

    return (
      <Container>
        <PageTitle hideHeading title={m.appData} />
        <Row>
          <Col className={style.wrapper}>
            <ApplicationEvents appId={appId} />
          </Col>
        </Row>
      </Container>
    )
  }
}
