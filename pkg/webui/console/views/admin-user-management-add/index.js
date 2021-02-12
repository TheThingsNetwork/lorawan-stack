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
import { push } from 'connected-react-router'

import PageTitle from '@ttn-lw/components/page-title'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import withRequest from '@ttn-lw/lib/components/with-request'

import UserDataForm from '@console/components/user-data-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'

import { createUser } from '@console/store/actions/users'
import { getIsConfiguration } from '@console/store/actions/identity-server'

import {
  selectPasswordRequirements,
  selectIsConfigurationFetching,
} from '@console/store/selectors/identity-server'

@connect(
  state => ({
    passwordRequirements: selectPasswordRequirements(state),
    fetching: selectIsConfigurationFetching(state),
  }),
  {
    createUser: attachPromise(createUser),
    navigateToList: () => push(`/admin/user-management`),
    getConfiguration: () => getIsConfiguration(),
  },
)
@withRequest(({ getConfiguration }) => getConfiguration(), ({ fetching }) => fetching)
@withBreadcrumb('admin.user-management.add', () => (
  <Breadcrumb path={`/admin/user-management/add`} content={sharedMessages.add} />
))
export default class UserManagementAdd extends Component {
  static propTypes = {
    createUser: PropTypes.func.isRequired,
    navigateToList: PropTypes.func.isRequired,
    passwordRequirements: PropTypes.passwordRequirements,
  }

  static defaultProps = {
    passwordRequirements: undefined,
  }

  @bind
  async onSubmit(values) {
    const { createUser } = this.props

    return createUser(values)
  }

  @bind
  onSubmitSuccess(response) {
    const { navigateToList } = this.props

    navigateToList()
  }

  render() {
    const { passwordRequirements } = this.props

    return (
      <Container>
        <PageTitle title={sharedMessages.userAdd} />
        <Row>
          <Col lg={8} md={12}>
            <UserDataForm
              passwordRequirements={passwordRequirements}
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
