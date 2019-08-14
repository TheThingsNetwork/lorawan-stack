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
import Query from 'query-string'
import { Redirect } from 'react-router-dom'
import { defineMessages } from 'react-intl'
import { Container, Col, Row } from 'react-grid-system'

import Message from '../../../lib/components/message'
import SafeInspector from '../../../components/safe-inspector'

const m = defineMessages({
  code: 'Your Authorization Code',
})

export default class Code extends React.Component {
  render() {
    const { location } = this.props
    const { query } = Query.parseUrl(location.search)

    if (!query.code) {
      return <Redirect to="/login" />
    }

    return (
      <Container>
        <Row>
          <Col xs={12}>
            <Message component="h2" content={m.code} />
            <SafeInspector data={query.code} initiallyVisible hideable={false} isBytes={false} />
          </Col>
        </Row>
      </Container>
    )
  }
}
