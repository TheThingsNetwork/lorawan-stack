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
import { defineMessages } from 'react-intl'
import { Container, Col, Row } from 'react-grid-system'

import PageTitle from '@ttn-lw/components/page-title'
import Collapse from '@ttn-lw/components/collapse'
import Overlay from '@ttn-lw/components/overlay'

import ProfileSettingsForm from '@account/containers/profile-settings-form'
import ChangePasswordForm from '@account/containers/change-password-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import useRequest from '@ttn-lw/lib/hooks/use-request'

import { getIsConfiguration } from '@account/store/actions/identity-server'

const m = defineMessages({
  profileEdit: 'Edit profile',
  generalSettingsDescription:
    'Change basic info such as your name, profile picture or email address.',
  changePasswordDescription: 'Set up a new password for your account.',
})

const ProfileSettings = () => {
  const [fetching, error] = useRequest(getIsConfiguration())

  if (Boolean(error)) {
    throw error
  }

  return (
    <Container>
      <Row>
        <Col lg={8} md={12}>
          <PageTitle title={m.profileEdit} />
          <Overlay after={350} visible={fetching} loading>
            <Collapse
              title={sharedMessages.generalSettings}
              description={m.generalSettingsDescription}
            >
              <ProfileSettingsForm />
            </Collapse>
            <Collapse
              title={sharedMessages.changePassword}
              description={m.changePasswordDescription}
            >
              <ChangePasswordForm />
            </Collapse>
          </Overlay>
        </Col>
      </Row>
    </Container>
  )
}

export default ProfileSettings
