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

import React, { useCallback } from 'react'
import { useSelector, useDispatch } from 'react-redux'
import { useNavigate } from 'react-router-dom'
import { Container, Col, Row } from 'react-grid-system'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import UserDataForm from '@console/components/user-data-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { createUser } from '@console/store/actions/users'
import { getIsConfiguration } from '@console/store/actions/identity-server'

import { selectPasswordRequirements } from '@console/store/selectors/identity-server'

const UserManagementAddInner = () => {
  const dispatch = useDispatch()
  const navigate = useNavigate()

  const passwordRequirements = useSelector(selectPasswordRequirements)
  const createUserAction = useCallback(values => dispatch(createUser(values)), [dispatch])

  const onSubmit = useCallback(values => createUserAction(values), [createUserAction])
  const onSubmitSuccess = useCallback(() => navigate('/admin-panel/user-management'), [navigate])

  return (
    <Container>
      <PageTitle title={sharedMessages.userAdd} />
      <Row>
        <Col>
          <UserDataForm
            passwordRequirements={passwordRequirements}
            onSubmit={onSubmit}
            onSubmitSuccess={onSubmitSuccess}
          />
        </Col>
      </Row>
    </Container>
  )
}

const UserManagementAdd = () => {
  useBreadcrumbs(
    'admin-panel.user-management.add',
    <Breadcrumb path={`/admin-panel/user-management/add`} content={sharedMessages.add} />,
  )

  return (
    <RequireRequest requestAction={getIsConfiguration()}>
      <UserManagementAddInner />
    </RequireRequest>
  )
}

export default UserManagementAdd
