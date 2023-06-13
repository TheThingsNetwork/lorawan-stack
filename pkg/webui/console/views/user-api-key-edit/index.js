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
import { Container, Col, Row } from 'react-grid-system'
import { useSelector } from 'react-redux'
import { useParams } from 'react-router-dom'

import { USER } from '@console/constants/entities'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import PageTitle from '@ttn-lw/components/page-title'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import { ApiKeyEditForm } from '@console/containers/api-key-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getUsersRightsList } from '@console/store/actions/users'
import { getApiKey } from '@console/store/actions/api-keys'

import { selectUserId } from '@account/store/selectors/user'

const UserApiKeyEditInner = () => {
  const userId = useSelector(selectUserId)
  const { apiKeyId } = useParams()

  useBreadcrumbs(
    'usr.single.api-keys.add',
    <Breadcrumb path={`/users/api-keys/edit/${apiKeyId}`} content={sharedMessages.add} />,
  )

  return (
    <Container>
      <PageTitle title={sharedMessages.addApiKey} />
      <Row>
        <Col lg={8} md={12}>
          <ApiKeyEditForm entity={USER} entityId={userId} />
        </Col>
      </Row>
    </Container>
  )
}

const UserApiKeyEdit = () => {
  const userId = useSelector(selectUserId)
  const { apiKeyId } = useParams()
  return (
    <RequireRequest
      requestAction={[getUsersRightsList(userId), getApiKey('users', userId, apiKeyId)]}
    >
      <UserApiKeyEditInner />
    </RequireRequest>
  )
}

export default UserApiKeyEdit
