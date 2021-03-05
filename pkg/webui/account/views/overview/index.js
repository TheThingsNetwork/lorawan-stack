// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import { useSelector } from 'react-redux'
import { Container, Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import PageTitle from '@ttn-lw/components/page-title'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import ProfileCard from '@account/containers/profile-card'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { selectConsoleUrl } from '@account/lib/selectors/app-config'

import { selectUser } from '@account/store/selectors/user'

import style from './overview.styl'

const m = defineMessages({
  accountAppInfoTitle: 'Welcome, {userId}! 👋',
  accountAppInfoMessage: `<p>You have successfully logged into the Account App. The Account App is the official user account management application of The Things Stack. In the near future, you will additionally be able to use this application to</p>
<ul><li>Manage your active sessions</li><li>Manage your OAuth authorizations</li></ul>
  `,
  accountAppConsoleInfo:
    'If you wish to manage your applications, end devices and/or gateways, you can use the button below to head over to the Console.',
  goToConsole: 'Go to the Console',
})

const consoleUrl = selectConsoleUrl()

const Overview = () => {
  const {
    name: userName,
    ids: { user_id: userId },
  } = useSelector(selectUser)

  return (
    <Container>
      <Row>
        <Col className={style.top}>
          <PageTitle title={sharedMessages.overview} hideHeading />
          <ProfileCard />
        </Col>
      </Row>
      <Row justify="center">
        <Col sm={6}>
          <Message
            component="h1"
            content={m.accountAppInfoTitle}
            values={{ userId: userName || userId }}
          />
          <Message
            content={m.accountAppInfoMessage}
            values={{
              p: msg => <p>{msg}</p>,
              ul: msg => <ol key="list">{msg}</ol>,
              li: msg => <li>{msg}</li>,
            }}
          />
          <Message component="p" content={m.accountAppConsoleInfo} />
          <Button.AnchorLink href={consoleUrl} message={m.goToConsole} />
        </Col>
      </Row>
    </Container>
  )
}

export default Overview
