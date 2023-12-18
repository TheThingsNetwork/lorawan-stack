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
import { Container, Col, Row } from 'react-grid-system'

import PageTitle from '@ttn-lw/components/page-title'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import InviteForm from '@console/containers/invite-user-form'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { maySendInvites } from '@console/lib/feature-checks'

const InvitationSend = () => {
  useBreadcrumbs('admin-panel.user-management.invitations.send', [
    {
      path: `/admin-panel/user-management/invitations/send`,
      content: sharedMessages.sendInvitation,
    },
  ])

  return (
    <Require featureCheck={maySendInvites} otherwise={{ redirect: '/' }}>
      <Container>
        <PageTitle title={sharedMessages.invite} />
        <Row>
          <Col>
            <InviteForm />
          </Col>
        </Row>
      </Container>
    </Require>
  )
}

export default InvitationSend
