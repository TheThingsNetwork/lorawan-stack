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

import React from 'react'
import { Navigate, useSearchParams } from 'react-router-dom'
import { defineMessages } from 'react-intl'
import { Container, Col, Row } from 'react-grid-system'

import SafeInspector from '@ttn-lw/components/safe-inspector'
import Button from '@ttn-lw/components/button'
import PageTitle from '@ttn-lw/components/page-title'

import Message from '@ttn-lw/lib/components/message'

import { selectApplicationSiteTitle } from '@ttn-lw/lib/selectors/env'

import style from './code.styl'

const m = defineMessages({
  code: 'Authorization code',
  codeDescription: 'Your authorization code is:',
  backToAccount: 'Back to {siteTitle}',
})

const siteTitle = selectApplicationSiteTitle()

const Code = () => {
  const [searchParams] = useSearchParams()
  const code = searchParams.get('code')

  if (!code) {
    return <Navigate to="/" />
  }

  return (
    <Container>
      <Row>
        <Col lg={4} md={6} sm={12}>
          <PageTitle title={m.code} />
          <Message
            content={m.codeDescription}
            component="label"
            className={style.codeDescription}
          />
          <SafeInspector
            data={code}
            initiallyVisible
            hideable={false}
            isBytes={false}
            className={style.code}
          />
          <Button.Link
            to="/"
            icon="keyboard_arrow_left"
            message={{ ...m.backToAccount, values: { siteTitle } }}
          />
        </Col>
      </Row>
    </Container>
  )
}

export default Code
