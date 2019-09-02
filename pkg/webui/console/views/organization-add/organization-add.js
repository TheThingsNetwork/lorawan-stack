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
import bind from 'autobind-decorator'
import { Container, Col, Row } from 'react-grid-system'

import OrganizationForm from '../../components/organization-form'

import PropTypes from '../../../lib/prop-types'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import sharedMessages from '../../../lib/shared-messages'

import style from './organization-add.styl'

const initialValues = {
  ids: {
    organization_id: '',
  },
  name: '',
  description: '',
}

class Add extends React.Component {
  static propTypes = {
    createOrganization: PropTypes.func.isRequired,
    createOrganizationSuccess: PropTypes.func.isRequired,
  }

  state = {
    error: '',
  }

  @bind
  handleSubmitFailure(error) {
    this.setState({ error })
  }

  render() {
    const { createOrganization, createOrganizationSuccess } = this.props
    const { error } = this.state

    return (
      <Container>
        <Row className={style.wrapper}>
          <Col sm={12}>
            <IntlHelmet title={sharedMessages.addOrganization} />
            <Message component="h2" content={sharedMessages.addOrganization} />
          </Col>
          <Col sm={12} md={8} lg={8} xl={8}>
            <OrganizationForm
              error={error}
              onSubmit={createOrganization}
              onSubmitSuccess={createOrganizationSuccess}
              onSubmitFailure={this.handleSubmitFailure}
              initialValues={initialValues}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}

export default Add
