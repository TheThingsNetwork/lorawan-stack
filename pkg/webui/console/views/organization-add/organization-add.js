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
import { defineMessages } from 'react-intl'

import PageTitle from '../../../components/page-title'
import OrganizationForm from '../../components/organization-form'
import SubmitBar from '../../../components/submit-bar'
import SubmitButton from '../../../components/submit-button'
import Form from '../../../components/form'
import PropTypes from '../../../lib/prop-types'
import sharedMessages from '../../../lib/shared-messages'
import { getOrganizationId } from '../../../lib/selectors/id'
import withFeatureRequirement from '../../lib/components/with-feature-requirement'

import { mayCreateOrganizations } from '../../lib/feature-checks'

const initialValues = {
  ids: {
    organization_id: '',
  },
  name: '',
  description: '',
}

const m = defineMessages({
  createOrganization: 'Create Organization',
})

@withFeatureRequirement(mayCreateOrganizations, { redirect: '/organizations' })
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

  @bind
  handleSubmitSuccess(organization) {
    const { createOrganizationSuccess } = this.props
    const orgId = getOrganizationId(organization)

    createOrganizationSuccess(orgId)
  }

  render() {
    const { createOrganization } = this.props
    const { error } = this.state

    return (
      <Container>
        <PageTitle tall title={sharedMessages.addOrganization} />
        <Row>
          <Col md={10} lg={9}>
            <OrganizationForm
              error={error}
              onSubmit={createOrganization}
              onSubmitSuccess={this.handleSubmitSuccess}
              onSubmitFailure={this.handleSubmitFailure}
              initialValues={initialValues}
            >
              <SubmitBar>
                <Form.Submit message={m.createOrganization} component={SubmitButton} />
              </SubmitBar>
            </OrganizationForm>
          </Col>
        </Row>
      </Container>
    )
  }
}

export default Add
