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

import PropTypes from '../../../lib/prop-types'
import toast from '../../../components/toast'
import sharedMessages from '../../../lib/shared-messages'
import withRequest from '../../../lib/components/with-request'
import PageTitle from '../../../components/page-title'
import UserDataForm from '../../components/user-data-form'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'

import diff from '../../../lib/diff'
import { selectSelectedUser } from '../../store/selectors/users'
import { getUser, updateUser } from '../../store/actions/users'
import { attachPromise } from '../../store/actions/lib'

const m = defineMessages({
  updateSuccess: 'User updated successfully',
  updateFailure: 'There was a problem updating the user',
})

@connect(
  (state, props) => ({
    userId: props.match.params.userId,
    user: selectSelectedUser(state),
  }),
  {
    getUser,
    updateUser: attachPromise(updateUser),
  },
)
@withRequest(
  ({ userId, getUser }) => getUser(userId, ['name', 'primary_email_address', 'state']),
  ({ fetching, user }) => fetching || !Boolean(user),
)
@withBreadcrumb('admin.user-management.edit', function({ userId }) {
  return (
    <Breadcrumb
      path={`/admin/user-management/${userId}`}
      icon="edit"
      content={sharedMessages.edit}
    />
  )
})
export default class UserManagementEdit extends Component {
  static propTypes = {
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

  @bind
  onSubmitSuccess() {
    const { userId } = this.props
    toast({
      title: userId,
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  @bind
  onSubmitFailure() {
    const { userId } = this.props
    toast({
      title: userId,
      message: m.updateFailure,
      type: toast.types.ERROR,
    })
  }

  render() {
    const { user } = this.props
    return (
      <Container>
        <PageTitle title={sharedMessages.userEdit} />
        <Row>
          <Col lg={8} md={12}>
            <UserDataForm
              initialValues={user}
              error={null}
              onSubmit={this.onSubmit}
              onSubmitSuccess={this.onSubmitSuccess}
              onSubmitFailure={this.onSubmitFailure}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
