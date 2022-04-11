// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
import { Col, Row, Container } from 'react-grid-system'

import DataSheet from '@ttn-lw/components/data-sheet'

import DateTime from '@ttn-lw/lib/components/date-time'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { selectSelectedClient, selectSelectedClientId } from '@account/store/selectors/clients'

const OAuthClientOverview = props => {
  const {
    oauthClientId,
    oauthClient: { created_at, updated_at },
  } = props
  console.log(oauthClientId)
  console.log(created_at)
  const sheetData = [
    {
      header: sharedMessages.generalInformation,
      items: [
        { key: sharedMessages.oauthClientId, value: oauthClientId, type: 'code', sensitive: false },
        { key: sharedMessages.createdAt, value: <DateTime value={created_at} /> },
        { key: sharedMessages.updatedAt, value: <DateTime value={updated_at} /> },
      ],
    },
  ]

  return (
    <>
      {/* <div className={style.titleSection}>
        <Container>
          <IntlHelmet title={sharedMessages.overview} />
          <Row>
            <Col sm={12}>
              <ApplicationTitleSection appId={oauthClientId} />
            </Col>
          </Row>
        </Container>
      </div> */}
      <Container>
        <Row>
          <Col sm={12} lg={6}>
            <DataSheet data={sheetData} />
          </Col>
        </Row>
      </Container>
    </>
  )
}

OAuthClientOverview.propTypes = {
  oauthClientId: PropTypes.string.isRequired,
}

export default connect(state => {
  const oauthClientId = selectSelectedClientId(state)

  return {
    oauthClientId,
    oauthClient: selectSelectedClient(state),
  }
})(OAuthClientOverview)
