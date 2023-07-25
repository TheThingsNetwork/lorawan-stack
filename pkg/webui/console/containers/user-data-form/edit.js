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
import { useDispatch, useSelector } from 'react-redux'
import { useNavigate, useParams } from 'react-router-dom'
import { Container, Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import toast from '@ttn-lw/components/toast'
import PageTitle from '@ttn-lw/components/page-title'

import UserDataForm from '@console/components/user-data-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import diff from '@ttn-lw/lib/diff'
import { getUserId } from '@ttn-lw/lib/selectors/id'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { updateUser, deleteUser } from '@console/store/actions/users'

import { selectSelectedUser } from '@console/store/selectors/users'

const m = defineMessages({
  updateSuccess: 'User updated',
  deleteSuccess: 'User deleted',
})

const UserDataFormEdit = () => {
  const dispatch = useDispatch()
  const navigate = useNavigate()
  const { userId } = useParams()
  const user = useSelector(selectSelectedUser)

  const wrappedUpdateUser = attachPromise(updateUser)
  const wrappedDeleteUser = attachPromise(deleteUser)

  const onSubmit = useCallback(
    values => {
      const patch = diff(user, values)
      const submitPatch = Object.keys(patch).length !== 0 ? patch : user
      return dispatch(wrappedUpdateUser(userId, submitPatch))
    },
    [user, userId, wrappedUpdateUser, dispatch],
  )

  const onSubmitSuccess = useCallback(response => {
    const userId = getUserId(response)
    toast({
      title: userId,
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
  }, [])

  const onDelete = useCallback(
    shouldPurge => dispatch(wrappedDeleteUser(userId, { purge: shouldPurge })),
    [userId, wrappedDeleteUser, dispatch],
  )

  const onDeleteSuccess = useCallback(() => {
    toast({
      title: userId,
      message: m.deleteSuccess,
      type: toast.types.SUCCESS,
    })

    navigate('../../')
  }, [userId, navigate])

  return (
    <Container>
      <PageTitle title={sharedMessages.userEdit} />
      <Row>
        <Col>
          <UserDataForm
            update
            initialValues={user}
            onSubmit={onSubmit}
            onSubmitSuccess={onSubmitSuccess}
            onDelete={onDelete}
            onDeleteSuccess={onDeleteSuccess}
          />
        </Col>
      </Row>
    </Container>
  )
}

export default UserDataFormEdit
