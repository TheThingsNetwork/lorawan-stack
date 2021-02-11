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

import React, { Component } from 'react'
import { connect } from 'react-redux'
import { Container, Col, Row } from 'react-grid-system'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import { replace } from 'connected-react-router'

import toast from '@ttn-lw/components/toast'
import PageTitle from '@ttn-lw/components/page-title'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import withRequest from '@ttn-lw/lib/components/with-request'

import UserDataForm from '@console/components/user-data-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import diff from '@ttn-lw/lib/diff'
import { getUserId } from '@ttn-lw/lib/selectors/id'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { getUser, updateUser, deleteUser } from '@console/store/actions/users'

import { selectSelectedUser } from '@console/store/selectors/users'

const m = defineMessages({
  updateSuccess: 'User updated',
  deleteSuccess: 'User deleted',
})

@connect(
  (state, props) => ({
    userId: props.match.params.userId,
    user: selectSelectedUser(state),
  }),
  {
    getUser,
    updateUser: attachPromise(updateUser),
    deleteUser: attachPromise(deleteUser),
    navigateToList: () => replace(`/admin/user-management`),
  },
)
@withRequest(
  ({ userId, getUser }) => getUser(userId, ['name', 'primary_email_address', 'state', 'admin']),
  ({ fetching, user }) => fetching || !Boolean(user),
)
@withBreadcrumb('admin.user-management.edit', function ({ userId }) {
  return <Breadcrumb path={`/admin/user-management/${userId}`} content={sharedMessages.edit} />
})
export default class UserManagementEdit extends Component {
  static propTypes = {
    deleteUser: PropTypes.func.isRequired,
    navigateToList: PropTypes.func.isRequired,
    updateUser: PropTypes.func.isRequired,
    user: PropTypes.user.isRequired,
    userId: PropTypes.string.isRequired,
  }

  @bind
  onSubmit(values) {
    const { userId, user, updateUser } = this.props
    const patch = diff(user, values)

    return updateUser(userId, patch)
  }

  onSubmitSuccess(response) {
    const userId = getUserId(response)
    toast({
      title: userId,
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  @bind
  onDelete() {
    const { userId, deleteUser } = this.props

    return deleteUser(userId)
  }

  @bind
  onDeleteSuccess() {
    const { userId, navigateToList } = this.props

    toast({
      title: userId,
      message: m.deleteSuccess,
      type: toast.types.SUCCESS,
    })

    navigateToList()
  }

  render() {
    const { user } = this.props
    return (
      <Container>
        <PageTitle title={sharedMessages.userEdit} />
        <Row>
          <Col lg={8} md={12}>
            <UserDataForm
              update
              initialValues={user}
              onSubmit={this.onSubmit}
              onSubmitSuccess={this.onSubmitSuccess}
              onDelete={this.onDelete}
              onDeleteSuccess={this.onDeleteSuccess}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
