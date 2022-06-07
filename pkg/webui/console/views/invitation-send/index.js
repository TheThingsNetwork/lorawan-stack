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
import { Container, Col, Row } from 'react-grid-system'
import { connect } from 'react-redux'
import { push } from 'connected-react-router'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import InviteForm from '@console/components/invite-user-form'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'

import { maySendInvites } from '@console/lib/feature-checks'

import { sendInvite } from '@console/store/actions/user'

const InvitationSend = props => {
  const { navigateToList, sendInvite } = props

  useBreadcrumbs(
    'admin.user-management.invitations.add',
    <Breadcrumb path={`/admin/user-management/invitations/add`} content={sharedMessages.add} />,
  )

  const onSubmit = React.useCallback(async email => sendInvite(email), [sendInvite])

  const onSubmitSuccess = React.useCallback(() => navigateToList(), [navigateToList])

  return (
    <Require featureCheck={maySendInvites} otherwise={{ redirect: '/' }}>
      <Container>
        <PageTitle title={sharedMessages.invite} />
        <Row>
          <Col lg={8} md={12}>
            <InviteForm onSubmit={onSubmit} onSubmitSuccess={onSubmitSuccess} />
          </Col>
        </Row>
      </Container>
    </Require>
  )
}

InvitationSend.propTypes = {
  navigateToList: PropTypes.func.isRequired,
  sendInvite: PropTypes.func.isRequired,
}

export default connect(null, {
  sendInvite: email => attachPromise(sendInvite(email)),
  navigateToList: () => push(`/admin/user-management`),
})(InvitationSend)
