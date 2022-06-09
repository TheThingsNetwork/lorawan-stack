// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import { Container, Row, Col } from 'react-grid-system'

import PageTitle from '@ttn-lw/components/page-title'

import WebhooksTable from '@console/containers/webhooks-table'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const ApplicationWebhooksList = props => {
  const { appId } = props.match.params

  return (
    <Container>
      <PageTitle title={sharedMessages.webhooks} hideHeading />
      <Row>
        <Col>
          <WebhooksTable appId={appId} />
        </Col>
      </Row>
    </Container>
  )
}

ApplicationWebhooksList.propTypes = {
  match: PropTypes.match.isRequired,
}

export default ApplicationWebhooksList
